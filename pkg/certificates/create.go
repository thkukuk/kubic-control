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

package certificates

import (
	"os"
	"fmt"

        "github.com/spf13/cobra"
)

func CreateCertsCmd() *cobra.Command {
        var subCmd = &cobra.Command {
                Use:   "create <user>",
                Short: "Cerate certificate for an user",
                Run: createCerts,
                Args: cobra.ExactArgs(1),
        }

        return subCmd
}

func createCerts (cmd *cobra.Command, args []string) {
	user := args[0]

	err := CreateUser(PKI_dir, user)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating certificate for user '%s': %v\n",
			user, err)
		return
	}
	err = SignUser(PKI_dir, user)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error signing certificate for user '%s': %v\n",
			user, err)
		return
	}
	fmt.Printf("Signed certificates for user '%s' created in '%s'.\n",
		user, PKI_dir)
}
