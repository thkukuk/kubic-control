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

package rbac

import (
	"os"
	"fmt"
	"strings"

        "github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

func AddAccountCmd() *cobra.Command {
        var subCmd = &cobra.Command {
                Use:   "add <role> <user>",
                Short: "Add user account to a role",
                Run: addAccount,
                Args: cobra.ExactArgs(2),
        }

        return subCmd
}

func addAccount (cmd *cobra.Command, args []string) {
	role := args[0]
	user := args[1]
	entry := ""

	cfg, err := ini.LooseLoad("/usr/share/defaults/kubicd/rbac.conf", "/etc/kubicd/rbac.conf")
        if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot load rbac.conf: %v\n", err)
		os.Exit(1)
	}

	if !cfg.Section("").HasKey(role) {
		fmt.Printf("Adding new role: '%s'\n", role)
	} else {
		entry = cfg.Section("").Key(role).String()
	}
	userList := strings.Split(entry, ",")
        for i := range userList {
                if user == strings.TrimSpace(userList[i]) {
			fmt.Printf("User already part of '%s'\n", role)
                        return
                }
        }
	if len(entry) > 0 {
		entry = entry + "," + user
	} else {
		entry = user
	}
	wcfg, werr := ini.LooseLoad("/etc/kubicd/rbac.conf")
	if werr != nil {
		fmt.Fprintf(os.Stderr, "Cannot open /etc/kubicd/rbac.conf: %v\n",
			werr)
		os.Exit(1)
	}
	wcfg.Section("").Key(role).SetValue(entry)
	werr = wcfg.SaveTo("/etc/kubicd/rbac.conf")
	if werr != nil {
		fmt.Fprintf(os.Stderr, "Writing rbac.conf failed: %v\n", werr)
		os.Exit (1)
	}
}
