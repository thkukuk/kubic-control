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

        log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	pb "github.com/thkukuk/kubic-control/api"
)

func ListNodesCmd() *cobra.Command {
        var subCmd = &cobra.Command {
                Use:   "list",
                Short: "List all reachable worker nodes",
                Run: listNodes,
		Args: cobra.ExactArgs(0),
	}

	return subCmd
}

func listNodes(cmd *cobra.Command, args []string) {
	// Set up a connection to the server.
	conn, err := CreateConnection()
	if err != nil {
		return
	}
	defer conn.Close()

	c := pb.NewKubeadmClient(conn)

	// var deadlineMin = flag.Int("deadline_min", 10, "Default deadline in minutes.")
	// clientDeadline := time.Now().Add(time.Duration(*deadlineMin) * time.Minute)
	// ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	r, err := c.ListNodes(ctx, &pb.Empty{})
	if err != nil {
		log.Errorf("could not initialize: %v", err)
		return
	}
	if r.Success {
		fmt.Printf("Reacheable nodes: ")
		for i := range r.Node {
			if i == 0 {
				fmt.Print(r.Node[i])
			} else {
				fmt.Print(", " + r.Node[i])
			}
                }
		fmt.Print("\n")
	} else {
		log.Errorf("Getting list of nodes failed: %s", r.Message)
	}
}
