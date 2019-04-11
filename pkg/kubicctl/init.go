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
	"os"
	"time"
	"flag"

        log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	pb "github.com/thkukuk/kubic-control/api"
)

var (
	defaultNetwork = "flannel"
)

func InitMasterCmd() *cobra.Command {
        var subCmd = &cobra.Command {
                Use:   "init",
                Short: "Initialize Kubernetes Master Node",
                Run: initMaster,
	}

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

	// Contact the server and print out its response.
	podNetwork := defaultNetwork
	if len(os.Args) > 1 {
		podNetwork = os.Args[1]
	}

	var deadlineMin = flag.Int("deadline_min", 10, "Default deadline in minutes.")
	clientDeadline := time.Now().Add(time.Duration(*deadlineMin) * time.Minute)
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.InitMaster(ctx, &pb.InitRequest{PodNetworking: podNetwork})
	if err != nil {
		log.Errorf("could not initialize: %v", err)
	}
	if r.Success {
		log.Printf("Kubernetes master was succesfully setup\n")
	} else {
		log.Errorf("Creating Kubernetes master failed: %s", r.Message)
	}
}
