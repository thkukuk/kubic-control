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
	"io"

	"github.com/spf13/cobra"
	pb "github.com/thkukuk/kubic-control/api"
)

var (
	podNetwork = "weave"
	adv_addr = ""
	multiMaster = ""
)

func InitMasterCmd() *cobra.Command {
        var subCmd = &cobra.Command {
                Use:   "init",
                Short: "Initialize Kubernetes Master Node",
                Run: initMaster,
		Args: cobra.ExactArgs(0),
	}

        subCmd.PersistentFlags().StringVar(&multiMaster, "multi-master", multiMaster, "Setup multimaster cluster, argument needs to be the DNS name of the load balancer")
        subCmd.PersistentFlags().StringVar(&podNetwork, "pod-network", podNetwork, "pod network, valid values are 'cilium', 'flannel' or 'weave'")
        subCmd.PersistentFlags().StringVar(&adv_addr, "adv-addr", adv_addr, "IP address the API Server will advertise it's listening on")

	return subCmd
}

func initMaster(cmd *cobra.Command, args []string) {
	// Set up a connection to the server.

	conn, err := CreateConnection()
	if err != nil {
		return
	}
	defer conn.Close()

	client := pb.NewKubeadmClient(conn)

	var deadlineMin = flag.Int("deadline_min", 10, "Default deadline in minutes.")
	clientDeadline := time.Now().Add(time.Duration(*deadlineMin) * time.Minute)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	fmt.Print ("Initializing kubernetes master can take several minutes, please be patient.\n")
	stream, err := client.InitMaster(ctx, &pb.InitRequest{PodNetworking: podNetwork, AdvAddr: adv_addr, MultiMaster: multiMaster})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize: %v\n", err)
		return
	}
	for {
		r, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			if r == nil {
				fmt.Fprintf(os.Stderr, "Creating Kubernetes master failed: %v\n",  err)
			} else {
				fmt.Fprintf(os.Stderr, "Creating Kubernetes master failed: %s\n%v\n", r.Message, err)
			}
			return
		}
		fmt.Printf("%s\n", r.Message)
	}
}
