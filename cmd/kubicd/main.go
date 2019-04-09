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
	"time"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
	"gopkg.in/ini.v1"
	log "github.com/sirupsen/logrus"
	"github.com/thkukuk/kubic-control/pkg/kubeadm"
	pb "github.com/thkukuk/kubic-control/api"
)

var (
	version = "unreleased"
	port = ":7148"
	crt = "certs/KubicD.crt"
	key = "certs/KubicD.key"
	ca = "certs/Kubic-Control.crt"
	rbac, err = ini.LooseLoad("/usr/share/defaults/kubicd/rbac.conf", "/etc/kubicd/rbac.conf")
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

func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no peer found")
	}
	tlsAuth, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unexpected peer transport credentials")
	}
	if len(tlsAuth.State.VerifiedChains) == 0 || len(tlsAuth.State.VerifiedChains[0]) == 0 {
		return nil, status.Error(codes.Unauthenticated, "could not verify peer certificate")
	}
	// Check subject common name against configured username
	log.Info(tlsAuth.State.VerifiedChains[0][0].Subject.CommonName)

	start := time.Now()
	// Calls the handler
	h, err := handler(ctx, req)

	log.Infof("Function: %s, Caller: %s, Duration: %s, Error: %v",
		info.FullMethod,
		tlsAuth.State.VerifiedChains[0][0].Subject.CommonName,
		time.Since(start), err)

	return h, err
}

func main() {
        log.Infof("Kubic Daemon: %s", version)

	// Load the certificates from disk
	certificate, err := tls.LoadX509KeyPair(crt, key)
	if err != nil {
		log.Fatalf("could not load server key pair: %s", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(ca)
	if err != nil {
		log.Fatalf("could not read ca certificate: %s", err)
	}

	// Append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatal("failed to append client certs")
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create the TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	})

	s := grpc.NewServer(grpc.Creds(creds),
		grpc.UnaryInterceptor(AuthInterceptor))
	pb.RegisterKubeadmServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
