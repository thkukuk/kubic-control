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
	"time"
	"fmt"
	"io"
	"os"

        log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	pb "github.com/thkukuk/kubic-control/api"
)

var (
	nodeType = "worker"
)

func AddNodeCmd() *cobra.Command {
        var subCmd = &cobra.Command {
                Use:   "add <node>",
                Short: "Add new nodes to cluster",
                Run: addNode,
		Args: cobra.ExactArgs(1),
	}

	subCmd.PersistentFlags().StringVar(&nodeType, "type", nodeType, "type of node, valid values are 'worker' or 'master'")

	return subCmd
}

func addNode(cmd *cobra.Command, args []string) {
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

	stream, err := client.AddNode(ctx, &pb.AddNodeRequest{NodeNames: nodes, Type: nodeType})
	if err != nil {
		log.Errorf("could not initialize: %v", err)
		return
	}

	for {
                r, err := stream.Recv()
                if err == io.EOF {
                        break
                }
                if err != nil {
                        if r == nil {
                                fmt.Fprintf(os.Stderr, "Adding node %s failed: %v\n", nodes, err)
                        } else {
                                fmt.Fprintf(os.Stderr, "Adding node %s failed: %s\n%v\n", r.Message, err)
                        }
                        os.Exit(1)
                }
                if (r.Success != true) {
                        fmt.Fprintf(os.Stderr, "%s\n", r.Message)
                        os.Exit(1)
                } else {
                        fmt.Printf("%s\n", r.Message)
                }
        }

	fmt.Print("Node(s) successfully added\n")
}
