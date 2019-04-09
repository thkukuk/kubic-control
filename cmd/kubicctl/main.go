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

// kubicctl implements a client for KubicAdmin service.
package main

import (
	"context"
	"os"
	"time"
	"flag"

        log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	pb "github.com/thkukuk/kubic-control/api"
)

const (
	address     = "localhost:50051"
	defaultNetwork = "flannel"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
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
		log.Fatalf("could not initialize: %v", err)
	}
	if r.Success {
		log.Printf("Kubernetes master was succesfully setup\n")
	} else {
		log.Warnf("Creating Kubernetes master failed: %s", r.Message)
	}
}
