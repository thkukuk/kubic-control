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
	"flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	pb "github.com/thkukuk/kubic-control/api"
)

var (
	podNetwork = "flannel"
)

func InitMasterCmd() *cobra.Command {
        var subCmd = &cobra.Command {
                Use:   "init",
                Short: "Initialize Kubernetes Master Node",
                Run: initMaster,
		Args: cobra.ExactArgs(0),
	}

        subCmd.PersistentFlags().StringVar(&podNetwork, "pod-network", podNetwork, "pod network should be used")

	return subCmd
}

func initMaster(cmd *cobra.Command, args []string) {
	// Set up a connection to the server.

	conn, err := CreateConnection()
	if err != nil {
		return
	}
	defer conn.Close()

	c := pb.NewKubeadmClient(conn)

	var deadlineMin = flag.Int("deadline_min", 10, "Default deadline in minutes.")
	clientDeadline := time.Now().Add(time.Duration(*deadlineMin) * time.Minute)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	fmt.Print ("Initializing kubernetes master can take several minutes, please be patient.\n")
	r, err := c.InitMaster(ctx, &pb.InitRequest{PodNetworking: podNetwork})
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not initialize: %v", err)
		return
	}
	if r.Success {
		fmt.Printf("Kubernetes master was succesfully setup\n")
	} else {
		fmt.Fprintf(os.Stderr, "Creating Kubernetes master failed: %s", r.Message)
	}
}
