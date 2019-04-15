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
        "github.com/spf13/cobra"
)

var (
	PKI_dir = "/etc/kubicd/pki"
)

func CertificatesCmd() *cobra.Command {
        var subCmd = &cobra.Command {
		Use:   "certificates",
                Short: "Manage certificates for kubicd/kubicctl communication",
        }

	subCmd.PersistentFlags().StringVar(&PKI_dir, "pki-dir", PKI_dir, "PKI directory to find and store certificates")


        subCmd.AddCommand(
                InitializeCertsCmd(),
        )

        return subCmd
}
