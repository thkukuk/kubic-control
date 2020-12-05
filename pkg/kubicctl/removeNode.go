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

func RemoveNodeCmd() *cobra.Command {
	var subCmd = &cobra.Command{
		Use:   "remove <node>",
		Short: "Remove node from cluster",
		Run:   removeNode,
		Args:  cobra.ExactArgs(1),
	}

	return subCmd
}

func removeNode(cmd *cobra.Command, args []string) {
	// Set up a connection to the server.

	nodes := args[0]

	conn, err := CreateConnection()
	if err != nil {
		return
	}
	defer conn.Close()

	client := pb.NewKubeadmClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	stream, err := client.RemoveNode(ctx, &pb.RemoveNodeRequest{NodeNames: nodes})
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
				fmt.Fprintf(os.Stderr, "Removing node %s failed: %v\n", nodes, err)
			} else {
				fmt.Fprintf(os.Stderr, "Removing node %s failed: %s\n%v\n", r.Message, err)
			}
			os.Exit(1)
		}
		if r.Success != true {
			fmt.Fprintf(os.Stderr, "%s\n", r.Message)
		} else {
			fmt.Printf("%s\n", r.Message)
		}
	}

	fmt.Printf("Please make sure to reboot the Nodes before re-using them.\n")
}
