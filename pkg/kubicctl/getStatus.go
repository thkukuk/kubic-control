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

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	pb "github.com/thkukuk/kubic-control/api"
)

func GetStatusCmd() *cobra.Command {
	var subCmd = &cobra.Command{
		Use:   "status",
		Short: "Status of current deployment",
		Run:   getStatus,
		Args:  cobra.ExactArgs(0),
	}

	return subCmd
}

func getStatus(cmd *cobra.Command, args []string) {
	// Set up a connection to the server.
	conn, err := CreateConnection()
	if err != nil {
		return
	}
	defer conn.Close()

	client := pb.NewKubeadmClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	stream, err := client.GetStatus(ctx, &pb.Empty{})
	if err != nil {
		log.Errorf("could not initialize: %v", err)
		return
	}

	fmt.Printf("Kubicctl version %s\n", Version)
	for {
		r, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			if r == nil {
				fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "ERROR: %s\n%v\n", r.Message, err)
			}
			os.Exit(1)
		}
		fmt.Printf("%s\n", r.Message)
	}
}
