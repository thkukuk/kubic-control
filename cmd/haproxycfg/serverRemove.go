// Copyright 2019 Thorsten Kukuk
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"io/ioutil"
	"strings"
	"fmt"
	"os"

        "github.com/spf13/cobra"
        "github.com/thkukuk/kubic-control/pkg/tools"
)

func ServerRemoveCmd() *cobra.Command {
        var subCmd = &cobra.Command {
                Use:   "remove <server>",
                Short: "Remove server from k8s-api backend entry",
                Run: serverRemove,
                Args: cobra.ExactArgs(1),
        }

	subCmd.PersistentFlags().StringVar(&OutputDir, "dir", OutputDir, "Directory, in which haproxy.cfg should be written")

        return subCmd
}

func serverRemove (cmd *cobra.Command, args []string) {

	oldApiserver := args[0]

	if len(OutputDir) > 0 && OutputDir[len(OutputDir)-1:] != "/" {
		OutputDir = OutputDir + "/"
	}

	// if the haproxy.cfg file does not exist or the force option is
	// given, creae new haproxy.cfg template
	found, _ := Exists(OutputDir + "haproxy.cfg")
	if !found {
		fmt.Fprintf(os.Stderr, "File not found: \"" + OutputDir + "haproxy.cfg\"")
		os.Exit(1)
	}

	/* Search for k8s-api backend and add our server entry */

	/* ioutil.ReadFile returns []byte, error */
	data, err := ioutil.ReadFile(OutputDir + "haproxy.cfg")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading \"" + OutputDir + "haproxy.cfg\": %v", err)
		os.Exit(1)
	}
	file := string(data)
	// Remove trailing \n to avoid additional new line
	file = strings.TrimSuffix(file, "\n")
	temp := strings.Split(file, "\n")

	var newContent []string
	modified := false
	found = false
	for _, item := range temp {
		if found == true &&
			strings.Contains(item, "server apiserver") {
			/* we are in the ackend k8s.api block and found an
                             apiserver entry, check the name */
			entry := strings.Fields(item)
			s := strings.TrimSuffix(entry[2], ":6443")
			if s == oldApiserver {
				// found entry, don't add it again
				modified = true
			} else {
				newContent = append(newContent, item)
			}
		} else if found == true && len(strings.TrimSpace(item)) == 0 {
			/* we are in the backend k8s-api block and found an empty line,
                             so are at the end of the block. Add now the server entries */
			found = false
			newContent = append(newContent, item)
		}else {
			if strings.HasPrefix(item, "backend") &&
				strings.Contains(item, "k8s-api") {
				found = true;
			}
			newContent = append (newContent, item)
		}
	}

	if !modified {
		/* we have not modified the file, something did go wrong, don't overwrite
                     existing file */
		fmt.Fprintf(os.Stderr, "Couldn't find server entry for '" + oldApiserver + "' in  \"" + OutputDir + "haproxy.cfg\"\n")
		os.Exit(1)
	}

	// XXX create backup of old file
	f, err := os.Create(OutputDir + "haproxy.cfg")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create \"" + OutputDir + "haproxy.cfg\": %v", err)
		os.Exit(1)
	}

	for _, item := range newContent {
		_, err = f.WriteString(item + "\n")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Wrting to \"" + OutputDir + "haproxy.cfg\" failed: %v", err)
			os.Exit(1)
		}

	}
	if err := f.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Closing \"" + OutputDir + "haproxy.cfg\" failed: %v", err)
		os.Exit(1)
	}

	set_perm (OutputDir + "haproxy.cfg")
	fmt.Printf("haproxy.cfg adjusted\n")
	success, message := tools.ExecuteCmd("systemctl", "restart", "haproxy")
	if !success {
		fmt.Fprintf(os.Stderr, "Error restarting haproxy: %s\n",
			message)
	} else {
		fmt.Print("haproxy restarted\n")
	}
}
