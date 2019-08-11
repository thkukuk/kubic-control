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
	"os/user"
	"strconv"

        "github.com/spf13/cobra"
	"github.com/thkukuk/kubic-control/pkg/tools"
)

var (
	force = false
)

func InitializeConfigCmd() *cobra.Command {
        var subCmd = &cobra.Command {
                Use:   "init <loadbalancer DNS name> <first master IP>",
                Short: "Create initial haproxy.cfg overwriting existing one",
                Run: initializeConfig,
                Args: cobra.ExactArgs(2),
        }

	subCmd.PersistentFlags().StringVar(&OutputDir, "dir", OutputDir, "Directory, in which haproxy.cfg should be written")
	subCmd.PersistentFlags().BoolVar(&force, "force", false, "force overwriting of existing haproxy.cfg")

        return subCmd
}

// exists returns whether the given file or directory exists
func Exists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return true, err
}

func set_perm(path string) error {
	err := os.Chmod(path, 0640)
	if err != nil {
		return err
	}

	if os.Getuid() != 0 {
		// don't change ownership if we are not root
		return nil
	}

	gr, err := user.LookupGroup("haproxy")
	if err != nil {
		// XXX if we don't have the group, do nothing
		return nil
	}

	gid, err := strconv.Atoi(gr.Gid)
	if err != nil {
		return err
	}
	return os.Chown(path, 0, gid)
}

func add_k8s_entry(f *os.File, lb_dns_name string, apiserver1 string) {
	_, err := f.WriteString("frontend k8s-api\n" +
		"    bind " + lb_dns_name + ":6443\n" +
		"    bind localhost:6443\n" +
		"    mode tcp\n" +
		"    option tcplog\n" +
		"    timeout client 125s\n" +
		"    default_backend k8s-api\n" +
		"\n" +
		"backend k8s-api\n" +
		"    mode tcp\n" +
		"    option tcp-check\n" +
		"    timeout server 125s\n" +
		"    balance roundrobin\n" +
		"    default-server inter 10s downinter 5s rise 2 fall 2 slowstart 60s maxconn 250 maxqueue 256 weight 100\n" +
		"    server apiserver1 " + apiserver1 + ":6443 check\n\n")
	if err != nil {
                fmt.Fprintf(os.Stderr, "Writing to haproxy.cfg failed: %v", err)
		os.Exit(1)
        }
}


func initializeConfig (cmd *cobra.Command, args []string) {

	lb_dns_name := args[0]
	apiserver1 := args[1]

	if len(OutputDir) > 0 && OutputDir[len(OutputDir)-1:] != "/" {
		OutputDir = OutputDir + "/"
	}

	// if the haproxy.cfg file does not exist or the force option is
	// given, creae new haproxy.cfg template
	found, _ := Exists(OutputDir + "haproxy.cfg")
	if !found || force {
		// Stop haproxy while we create the new config
		tools.ExecuteCmd("systemctl", "stop", "haproxy")
		f, err := os.Create(OutputDir + "haproxy.cfg")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create \"" + OutputDir + "haproxy.cfg\": %v", err)
			os.Exit(1)
		}

		_, err = f.WriteString("global\n" +
			"  log /dev/log daemon\n" +
			"  maxconn 32768\n" +
			"  chroot /var/lib/haproxy\n" +
			"  user haproxy\n" +
			"  group haproxy\n" +
			"  daemon\n" +
			"  stats socket /var/lib/haproxy/stats user haproxy group haproxy mode 0640 level operator\n" +
			"  tune.bufsize 32768\n" +
			"  tune.ssl.default-dh-param 2048\n" +
			"  ssl-default-bind-ciphers ALL:!aNULL:!eNULL:!EXPORT:!DES:!3DES:!MD5:!PSK:!RC4:!ADH:!LOW@STRENGTH\n" +
			"\n" +
			"defaults\n" +
			"  log     global\n" +
			"  mode    http\n" +
			"  option  log-health-checks\n" +
			"  option  log-separate-errors\n" +
			"  option  dontlog-normal\n" +
			"  option  dontlognull\n" +
			"  option  httplog\n" +
			"  option  socket-stats\n" +
			"  retries 3\n" +
			"  option  redispatch\n" +
			"  maxconn 10000\n" +
			"  timeout connect     5s\n" +
			"  timeout client     50s\n" +
			"  timeout server     50s\n" +
			"\n" +
			"listen stats\n" +
			"  bind localhost:80\n" +
			"  stats enable\n" +
			"  stats uri     /stats\n" +
			"  stats refresh 5s\n" +
			"\n")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Wrting to \"" + OutputDir + "haproxy.cfg\" failed: %v", err)
			os.Exit(1)
		}

		add_k8s_entry(f, lb_dns_name, apiserver1)

		if err := f.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Closing \"" + OutputDir + "haproxy.cfg\" failed: %v", err)
			os.Exit(1)
		}

		set_perm (OutputDir + "haproxy.cfg")
		fmt.Printf("haproxy.cfg created\n")
		success, message := tools.ExecuteCmd("systemctl", "enable", "--now", "haproxy")
		if !success {
			fmt.Fprintf(os.Stderr, "Error enabling and starting haproxy: %s\n",
				message)
		} else {
			fmt.Print("haproxy enabled and started\n")
		}
	} else {
		// File exists, we don't overwrite it, so remove existing
		// k8s-api frontend/backend and write them new.

		/* ioutil.ReadFile returns []byte, error */
		data, err := ioutil.ReadFile(OutputDir + "haproxy.cfg")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading \"" + OutputDir + "haproxy.cfg\": %v", err)
			os.Exit(1)
		}
		file := string(data)
		// Remove trailing \n to avoid additional new line
		file = strings.TrimSuffix(file, "\n")
		/* func Split(s, sep string) []string */
		temp := strings.Split(file, "\n")

		var newContent []string
		remove := false
		for _, item := range temp {
			if len(item) == 0  && remove {
				remove = false
			} else {
				if (strings.Contains(item, "frontend") ||
					strings.Contains(item, "backend")) &&
					strings.Contains(item, "k8s-api") {
					remove = true
				}
				if remove == false {
					newContent = append (newContent, item)
				}
			}
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
		add_k8s_entry(f, lb_dns_name, apiserver1)

		if err := f.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Closing \"" + OutputDir + "haproxy.cfg\" failed: %v", err)
			os.Exit(1)
		}

		set_perm (OutputDir + "haproxy.cfg")
		fmt.Printf("haproxy.cfg adjusted\n")
		// XXX check if service is enabled and running!
		success, message := tools.ExecuteCmd("systemctl", "restart", "haproxy")
		if !success {
			fmt.Fprintf(os.Stderr, "Error restarting haproxy: %s\n",
				message)
		} else {
			fmt.Print("haproxy restarted\n")
		}
	}
}
