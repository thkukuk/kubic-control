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
	"os"

        log "github.com/sirupsen/logrus"
        "github.com/spf13/cobra"
)
var (
        Version = "unreleased"
	OutputDir = "/etc/haproxy"
)

func ServerCmd() *cobra.Command {
        var subCmd = &cobra.Command {
                Use:   "server",
                Short: "Manage server entry of k8s-api backend",
        }

        subCmd.AddCommand(
                ServerAddCmd(),
        )

        return subCmd
}

func main() {
	rootCmd := &cobra.Command{
                Use:   "haproxycfg",
                Short: "Kubic haproxy.cfg configurator"}
	rootCmd.Version = Version
	rootCmd.AddCommand(
                VersionCmd(),
		InitializeConfigCmd(),
		ServerCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
        }
	os.Exit(0)
}
