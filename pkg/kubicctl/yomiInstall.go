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

func YomiInstallCmd() *cobra.Command {
        var subCmd = &cobra.Command {
                Use:   "finish <name>",
                Short: "Install a new node with yomi",
                Run: install,
		Args: cobra.ExactArgs(1),
	}

	return subCmd
}

func install(cmd *cobra.Command, args []string) {

	retval := 0
	node := args[0]

	// Set up a connection to the server.
	conn, err := CreateConnection()
	if err != nil {
		return
	}
	defer conn.Close()

	client := pb.NewYomiClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	stream, err := client.Install(ctx, &pb.InstallRequest{Saltnode: node})
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
                                fmt.Fprintf(os.Stderr, "Installing '%s' with yomi failed: %v\n",
					node, err)
                        } else {
                                fmt.Fprintf(os.Stderr, "Installing '%s' with yomi failed: %s\n%v\n",
					node, r.Message, err)
                        }
			retval = 1;
                } else {
			if (r.Success != true) {
				fmt.Fprintf(os.Stderr, "%s\n", r.Message)
			} else {
				fmt.Printf("%s\n", r.Message)
			}
		}
        }
	os.Exit(retval)
}
