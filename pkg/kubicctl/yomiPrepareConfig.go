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
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	pb "github.com/thkukuk/kubic-control/api"
)

var (
	arg_type        string
	arg_efi         bool
	arg_baremetal   bool
	arg_disk        string
	arg_repo        string
	arg_repo_update string
)

func YomiPrepareConfigCmd() *cobra.Command {
	var subCmd = &cobra.Command{
		Use:   "prepare <type> <name>",
		Short: "Prepare yomi configuration for node of this type",
		Run:   prepareConfig,
		Args:  cobra.ExactArgs(2),
	}

	subCmd.PersistentFlags().StringVar(&arg_type, "type", "", "Type of node: haproxy, master, worker")
	subCmd.PersistentFlags().StringVar(&arg_disk, "disk", "", "Disk device for installation")
	subCmd.PersistentFlags().StringVar(&arg_repo, "repo", "", "Repository to install from")
	subCmd.PersistentFlags().StringVar(&arg_repo_update, "update-repo", "", "Update repository to install from")
	subCmd.PersistentFlags().BoolVar(&arg_efi, "efi", false, "Machine has EFI firmware")
	subCmd.PersistentFlags().BoolVar(&arg_baremetal, "baremetal", false, "Machine is bare metal")

	return subCmd
}

func prepareConfig(cmd *cobra.Command, args []string) {

	retval := 0
	nodeType := args[0]
	node := args[1]

	// Set up a connection to the server.
	conn, err := CreateConnection()
	if err != nil {
		return
	}
	defer conn.Close()

	client := pb.NewYomiClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// XXX efi and baremetal are missing
	stream, err := client.PrepareConfig(ctx, &pb.PrepareConfigRequest{Saltnode: node, Type: nodeType,
		Disk: arg_disk, Repo: arg_repo, RepoUpdate: arg_repo_update})
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not initialize: %v", err)
		return
	}

	for {
		r, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			if r == nil {
				fmt.Fprintf(os.Stderr, "Create yomi configuration failed: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Create yomi configuration failed: %s\n%v\n", r.Message, err)
			}
			retval = 1
		} else {
			if r.Success != true {
				fmt.Fprintf(os.Stderr, "%s\n", r.Message)
			} else {
				fmt.Printf("%s\n", r.Message)
			}
		}
	}
	os.Exit(retval)
}
