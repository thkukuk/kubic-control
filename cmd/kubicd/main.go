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

package main

import (
	"context"
	"net"

	"google.golang.org/grpc"
	log "github.com/sirupsen/logrus"
	"github.com/thkukuk/kubic-control/pkg/kubeadm"
	pb "github.com/thkukuk/kubic-control/api"
)

var (
	version = "unreleased"
	port = ":50051"
)

type server struct{}

func (s *server) InitMaster(ctx context.Context, in *pb.InitRequest) (*pb.StatusReply, error) {
	log.Infof("Received: Kubernetes Version=%v, POD Network=%v",
		in.KubernetesVersion, in.PodNetworking)
	status, message := kubeadm.Init(in.PodNetworking, in.KubernetesVersion)
	return &pb.StatusReply{Success: status, Message: message}, nil
}

func (s *server) AddNode(ctx context.Context, in *pb.AddNodeRequest) (*pb.StatusReply, error) {
	log.Printf("Received: add node  %v", in.NodeName)
	return &pb.StatusReply{Success: true}, nil
}

func main() {
        log.Infof("Kubic Daemon: %s", version)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterKubeadmServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
