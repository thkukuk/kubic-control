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

package kubicctl

import (
	"os"
	"fmt"

        "github.com/spf13/cobra"
)

func InitializeCertsCmd() *cobra.Command {
        var subCmd = &cobra.Command {
                Use:   "initialize",
                Short: "Cerate CA, KubicD and admin certificates",
                Run: initializeCerts,
                Args: cobra.ExactArgs(0),
        }

        return subCmd
}

func initializeCerts (cmd *cobra.Command, args []string) {
	err := CreateCA(PKI_dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating CA: %v\n", err)
		return
	}
	err = CreateUser(PKI_dir, "KubicD")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating user 'KubicD': %v\n", err)
		return
	}
	err = SignUser(PKI_dir, "KubicD")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error signing user 'KubicD': %v\n", err)
		return
	}
	err = CreateUser(PKI_dir, "admin")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating user 'admin': %v\n", err)
		return
	}
	err = SignUser(PKI_dir, "admin")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error signing user 'admin': %v\n", err)
		return
	}
	fmt.Printf("All certificates and the CA are created and can be found in '%s'\n", PKI_dir)
}
