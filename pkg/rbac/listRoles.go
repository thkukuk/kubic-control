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

        "github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

func ListRolesCmd() *cobra.Command {
        var subCmd = &cobra.Command {
                Use:   "list",
                Short: "List roles and accounts",
                Run: listRoles,
                Args: cobra.ExactArgs(0),
        }

        return subCmd
}

func listRoles (cmd *cobra.Command, args []string) {
	cfg, err := ini.LooseLoad("/usr/etc/kubicd/rbac.conf", "/etc/kubicd/rbac.conf")
        if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot load rbac.conf: %v\n", err)
		os.Exit(1)
	}

	roleList := cfg.Section("").KeyStrings()
	for i := range roleList {
		entry := cfg.Section("").Key(roleList[i]).String()
		fmt.Printf("%s: %s\n", roleList[i], entry)
	}
}
