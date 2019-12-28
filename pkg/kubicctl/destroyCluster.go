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
	"os"
	"io"

	"github.com/spf13/cobra"
	pb "github.com/thkukuk/kubic-control/api"
)

func DestroyClusterCmd() *cobra.Command {
        var subCmd = &cobra.Command {
                Use:   "destroy-cluster",
                Short: "Destroy the whole cluster",
                Run: destroyCluster,
		Args: cobra.ExactArgs(0),
	}

	return subCmd
}

func destroyCluster(cmd *cobra.Command, args []string) {

	// Set up a connection to the server.
	conn, err := CreateConnection()
	if err != nil {
		return
	}
	defer conn.Close()

	client := pb.NewKubeadmClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	stream, err := client.RemoveNode(ctx, &pb.RemoveNodeRequest{NodeNames: "*"})
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
                                fmt.Fprintf(os.Stderr, "Removing all nodes failed: %v\n", err)
                        } else {
                                fmt.Fprintf(os.Stderr, "Removing node %s failed: %s\n%v\n", r.Message, err)
                        }
			os.Exit(1)
                }
		if (r.Success != true) {
			fmt.Fprintf(os.Stderr, "%s\n", r.Message)
		} else {
			fmt.Printf("%s\n", r.Message)
		}
        }

	fmt.Printf("All nodes removed, removing master...\n")

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	stream, err = client.DestroyMaster(ctx, &pb.Empty{})
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
                                fmt.Fprintf(os.Stderr, "Destroying master failed: %v\n", err)
                        } else {
                                fmt.Fprintf(os.Stderr, "Destroying master failed: %s\n%v\n", r.Message, err)
                        }
			os.Exit(1)
                }
		if (r.Success != true) {
			fmt.Fprintf(os.Stderr, "%s\n", r.Message)
		} else {
			fmt.Printf("%s\n", r.Message)
		}
        }

	fmt.Printf("Kubernetes cluster completly removed!\n")
}
