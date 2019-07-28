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
	"strings"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
	"gopkg.in/ini.v1"
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"
	"github.com/thkukuk/kubic-control/pkg/kubeadm"
	"github.com/thkukuk/kubic-control/pkg/certificate_server"
	pb "github.com/thkukuk/kubic-control/api"
)

var (
	Version = "unreleased"
	servername = "localhost"
	port = "7148"
	crtFile = "/etc/kubicd/pki/KubicD.crt"
	keyFile = "/etc/kubicd/pki/KubicD.key"
	caFile = "/etc/kubicd/pki/Kubic-Control-CA.crt"
	cfg, cfg_err = ini.LooseLoad("/usr/share/defaults/kubicd/kubicd.conf", "/etc/kubicd/kubicd.conf")
)

type kubeadm_server struct{}
type cert_server struct{}

// kubeadm API
func (s *kubeadm_server) InitMaster(in *pb.InitRequest, stream pb.Kubeadm_InitMasterServer) error {
	log.Infof("Received: Kubernetes Version=%v, POD Network=%v",
                in.KubernetesVersion, in.PodNetworking)
	return kubeadm.InitMaster(in, stream)
}

func (s *kubeadm_server) DestroyMaster(in *pb.Empty, stream pb.Kubeadm_DestroyMasterServer) error {
	log.Infof("Received: Destroy Master")
	return kubeadm.DestroyMaster(in, stream)
}

func (s *kubeadm_server) UpgradeKubernetes(in *pb.Empty, stream pb.Kubeadm_UpgradeKubernetesServer) error {
	log.Infof("Received: upgrade Kubernetes")
	return kubeadm.UpgradeKubernetes(in, stream)
}

func (s *kubeadm_server) RemoveNode(in *pb.RemoveNodeRequest, stream pb.Kubeadm_RemoveNodeServer) error {
	log.Printf("Received: remove node  %v", in.NodeNames)
	return kubeadm.RemoveNode(in, stream)
}

func (s *kubeadm_server) AddNode(ctx context.Context, in *pb.AddNodeRequest) (*pb.StatusReply, error) {
	log.Printf("Received: add node  %v", in.NodeNames)
	status, message := kubeadm.AddNode(in.NodeNames, in.Type)
	return &pb.StatusReply{Success: status, Message: message}, nil
}

func (s *kubeadm_server) RebootNode(ctx context.Context, in *pb.RebootNodeRequest) (*pb.StatusReply, error) {
	log.Printf("Received: reboot node  %v", in.NodeNames)
	status, message := kubeadm.RebootNode(in.NodeNames)
	return &pb.StatusReply{Success: status, Message: message}, nil
}

func (s *kubeadm_server) ListNodes(ctx context.Context, in *pb.Empty) (*pb.ListReply, error) {
	log.Printf("Received: list nodes")
	status, message, nodes := kubeadm.ListNodes()
	return &pb.ListReply{Success: status, Message: message, Node: nodes}, nil
}

func (s *kubeadm_server) FetchKubeconfig(ctx context.Context, in *pb.Empty) (*pb.StatusReply, error) {
	log.Printf("Received: fetch kubeconfig")
	status, message := kubeadm.FetchKubeconfig()
	return &pb.StatusReply{Success: status, Message: message}, nil
}

// Certificate API
func (s *cert_server) CreateCert(ctx context.Context, in *pb.CreateCertRequest) (*pb.CertificateReply, error) {
	log.Printf("Received: create certificate")
	status, message, key, crt := certificate.CreateCert(in)
	return &pb.CertificateReply{Success: status, Message: message, Key: key, Crt: crt}, nil
}


func rbacCheck(user string, function string) bool {

	rbac, rbac_err := ini.LooseLoad("/usr/share/defaults/kubicd/rbac.conf", "/etc/kubicd/rbac.conf")

	if rbac_err != nil {
		log.Error ("Error opening rbac config file: %v", rbac_err)
		return false
	}

	api := strings.TrimPrefix(function, "/api.")

	if !rbac.Section("").HasKey(api) {
		log.Errorf ("RBAC: no entry for '%s'", api)
		return false
	}
	value := rbac.Section("").Key(api).String()
	userList := strings.Split(value, ",")
	for i := range userList {
		if user == strings.TrimSpace(userList[i]) {
			return true
		}
	}

	log.Warnf("User '%s' wants access to function '%s', refused", user, function)

	return false
}

func AuthUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

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
	ok = rbacCheck(tlsAuth.State.VerifiedChains[0][0].Subject.CommonName, info.FullMethod)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "permission denied")
	}

	start := time.Now()
	// Calls the handler
	h, err := handler(ctx, req)

	log.Infof("Function: %s, Caller: %s, Duration: %s, Error: %v",
		info.FullMethod,
		tlsAuth.State.VerifiedChains[0][0].Subject.CommonName,
		time.Since(start), err)

	return h, err
}

func AuthStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

	p, ok := peer.FromContext(ss.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "no peer found")
	}
	tlsAuth, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return status.Error(codes.Unauthenticated, "unexpected peer transport credentials")
	}
	if len(tlsAuth.State.VerifiedChains) == 0 || len(tlsAuth.State.VerifiedChains[0]) == 0 {
		return status.Error(codes.Unauthenticated, "could not verify peer certificate")
	}
	// Check subject common name against configured username
	ok = rbacCheck(tlsAuth.State.VerifiedChains[0][0].Subject.CommonName, info.FullMethod)
	if !ok {
		return status.Error(codes.Unauthenticated, "permission denied")
	}

	start := time.Now()
	// Calls the handler
	err := handler(srv, ss)

	log.Infof("Function: %s, Caller: %s, Duration: %s, Error: %v",
		info.FullMethod,
		tlsAuth.State.VerifiedChains[0][0].Subject.CommonName,
		time.Since(start), err)

	return err
}


func loadConfigFile() {
	if cfg_err, ok := cfg_err.(*os.PathError); ok {
		log.Fatal(cfg_err)
	}

	if cfg.Section("global").HasKey("crtfile") {
		crtFile =  cfg.Section("global").Key("crtfile").String()
	}
	if cfg.Section("global").HasKey("keyfile") {
		keyFile =  cfg.Section("global").Key("keyfile").String()
	}
	if cfg.Section("global").HasKey("cafile") {
		caFile =  cfg.Section("global").Key("cafile").String()
	}
	if cfg.Section("global").HasKey("server") {
		servername =  cfg.Section("global").Key("server").String()
	}
	if cfg.Section("global").HasKey("port") {
		port =  cfg.Section("global").Key("port").String()
	}
}

func main() {
	loadConfigFile()

	rootCmd := &cobra.Command{
                Use:   "kubicd",
                Short: "Kubic Control  Daemon",
                Run:   kubicd,
	        Args: cobra.ExactArgs(0),
	}

	rootCmd.Version = Version
        rootCmd.PersistentFlags().StringVarP(&servername, "server", "s", servername, "Servername kubicd is listening on")
        rootCmd.PersistentFlags().StringVarP(&port, "port", "p", port, "Port on which kubicd is listening")
        rootCmd.PersistentFlags().StringVar(&crtFile, "crtfile", crtFile, "Certificate with the public key for the daemon")
        rootCmd.PersistentFlags().StringVar(&keyFile, "keyfile", keyFile, "Private key for the daemon")
        rootCmd.PersistentFlags().StringVar(&caFile, "cafile", caFile, "Certificate with the public key of the CA for the server certificate")

	if err := rootCmd.Execute(); err != nil {
                os.Exit (1)
        }
}

func kubicd(cmd *cobra.Command, args []string) {
        log.Infof("Kubic Daemon: %s", Version)

	// Create directory in /var/lib
	err := os.MkdirAll("/var/lib/kubic-control", os.ModePerm)
	if err != nil {
		log.Fatalf("Could not create '/var/lib/kubic-control' directory: %s", err)
	}

	// Load the certificates from disk
	certificate, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		log.Fatalf("Could not load server key pair: %s", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatalf("Could not read ca certificate: %s", err)
	}

	// Append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatal("Failed to append client certs")
	}

	lis, err := net.Listen("tcp", servername + ":" + port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create the TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	})

	s := grpc.NewServer(grpc.Creds(creds),
		grpc.StreamInterceptor(AuthStreamInterceptor),
		grpc.UnaryInterceptor(AuthUnaryInterceptor))

	pb.RegisterKubeadmServer(s, &kubeadm_server{})
	pb.RegisterCertificateServer(s, &cert_server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
