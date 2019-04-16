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

package kubeadm

import (
	"os"
	"strings"

	"gopkg.in/ini.v1"
	pb "github.com/thkukuk/kubic-control/api"
)

// exists returns whether the given file or directory exists
func exists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return true, err
}

func InitMaster(in *pb.InitRequest, stream pb.Kubeadm_InitMasterServer) error {
	arg_socket := "--cri-socket=/run/crio/crio.sock"
	arg_pod_network_cidr := ""
	arg_kubernetes_version := ""

	found, _ := exists ("/etc/kubernetes/manifests/kube-apiserver.yaml")
	if found == true {
		if err := stream.Send(&pb.StatusReply{Success: false, Message: "Seems like a kubernetes control-plane is already running. If not, please use \"kubeadm reset\" to clean up the system."}); err != nil {
			return err
		}
		return nil
	}
	found, _ = exists ("/etc/kubernetes/manifests/kube-scheduler.yaml")
	if found == true {
		if err := stream.Send(&pb.StatusReply{Success: false, Message: "Seems like a kubernetes control-plane is already running. If not, please use \"kubeadm reset\" to clean up the system"}); err != nil {
			return err
		}
		return nil
	}
	found, _ = exists ("/etc/kubernetes/manifests/etcd.yaml")
	if found == true {
		if err := stream.Send(&pb.StatusReply{Success: false, Message: "Seems like a kubernetes control-plane is already running. If not, please use \"kubeadm reset\" to clean up the system"}); err != nil {
			return err
		}
		return nil
	}

	success, message := ExecuteCmd("systemctl", "enable", "--now", "crio")
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
			return err
		}
		return nil
	}
	success, message = ExecuteCmd("systemctl", "enable", "--now", "kubelet")
	if success != true {
		ExecuteCmd("systemctl", "disable", "--now", "crio")
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
			return err
		}
		return nil
	}

	if (strings.EqualFold(in.PodNetworking, "flannel")) {
		arg_pod_network_cidr = "--pod-network-cidr=10.244.0.0/16"
	}
	if len (in.KubernetesVersion) > 0 {
		arg_kubernetes_version = "--kubernetes-version=" + in.KubernetesVersion
	} else {
		// No version given. Try to use kubeadm RPM version number.
		success, message := ExecuteCmd("rpm", "-q", "--qf", "'%{VERSION}'",  "kubernetes-kubeadm")
		if success == true {
			kubernetes_version := strings.Replace(message, "'", "", -1)
			arg_kubernetes_version = "--kubernetes-version="+kubernetes_version
		}
	}

	if err := stream.Send(&pb.StatusReply{Success: true, Message: "Initialize Kubernetes control-plane"}); err != nil {
		return err
	}
	success, message = ExecuteCmd("kubeadm", "init", arg_socket,
		arg_pod_network_cidr, arg_kubernetes_version)
	if success != true {
		ResetMaster()
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
			return err
		}
		return nil
	}

	// Setting up flannel
	if err := stream.Send(&pb.StatusReply{Success: true, Message: "Deploy flannel"}); err != nil {
		return err
	}
	success, message = ExecuteCmd("kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",  "apply", "-f", "https://raw.githubusercontent.com/coreos/flannel/bc79dd1505b0c8681ece4de4c0d86c5cd2643275/Documentation/kube-flannel.yml")
	if success != true {
		ResetMaster()
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
			return err
		}
		return nil
	}

	// Setting up kured
	if err := stream.Send(&pb.StatusReply{Success: true, Message: "Deploy Kubernetes Reboot Daemon (kured)"}); err != nil {
		return err
	}
	success, message = ExecuteCmd("kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",  "apply", "-f", "/usr/share/k8s-yaml/kured/kured.yaml")
	if success != true {
		ResetMaster()
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
			return err
		}
		return nil
	}
	// Configure transactional-update to inform kured
	ini.PrettyFormat = false
	ini.PrettyEqual = false
	cfg, err := ini.LooseLoad("/etc/transactional-update.conf")
	if err != nil {
		stream.Send(&pb.StatusReply{Success: true, Message: "Adjusting transactional-update to use kured for reboot failed.\nPlease ajdust /etc/transactional-update.conf yourself."})
	} else {
		cfg.Section("").Key("REBOOT_METHOD").SetValue("kured")
		cfg.SaveTo("/etc/transactional-update.conf")
	}

	if err := stream.Send(&pb.StatusReply{Success: true, Message: "Kubernetes master was succesfully setup."}); err != nil {
		return err
	}
	return nil
}
