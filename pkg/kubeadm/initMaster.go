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

	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
	pb "github.com/thkukuk/kubic-control/api"
	"github.com/thkukuk/kubic-control/pkg/tools"
	"github.com/thkukuk/kubic-control/pkg/deployment"
)

var (
	cilium_yaml = "/usr/share/k8s-yaml/cilium/cilium.yaml"
	flannel_yaml = "/usr/share/k8s-yaml/flannel/kube-flannel.yaml"
	weave_yaml = "/usr/share/k8s-yaml/weave/weave.yaml"
	kured_yaml = "/usr/share/k8s-yaml/kured/kured.yaml"
)

// update data in /var/lib/kubic-control
func update_cfg (file string, key string, value string) (error) {
        cfg, err := ini.LooseLoad("/var/lib/kubic-control/" + file)
        if err != nil {
                return err
        }

        cfg.Section("").Key(key).SetValue(value)
        err = cfg.SaveTo("/var/lib/kubic-control/" + file)
        if err != nil {
                return err
        }

        return nil
}

// exists returns whether the given file or directory exists
func exists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return true, err
}

func InitMaster(in *pb.InitRequest, stream pb.Kubeadm_InitMasterServer) error {
	arg_pod_network := in.PodNetworking

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

	// verify, that we got only a supported pod network
	if len(arg_pod_network) < 1 {
		arg_pod_network = "weave"
	}

	if strings.EqualFold(arg_pod_network, "weave") {
		found, _ = exists (weave_yaml)
		if found != true {
			if err := stream.Send(&pb.StatusReply{Success: false, Message: "weave-k8s-yaml is not installed!"}); err != nil {
				return err
			}
			return nil
		}
	} else if strings.EqualFold(arg_pod_network, "flannel") {
		found, _ = exists (flannel_yaml)
		if found != true {
			if err := stream.Send(&pb.StatusReply{Success: false, Message: "flannel-k8s-yaml is not installed!"}); err != nil {
				return err
			}
			return nil
		}
	} else if strings.EqualFold(arg_pod_network, "cilium") {
		found, _ = exists (cilium_yaml)
		if found != true {
			if err := stream.Send(&pb.StatusReply{Success: false, Message: "cilium-k8s-yaml is not installed!"}); err != nil {
				return err
			}
			return nil
		}
	} else {
		if err := stream.Send(&pb.StatusReply{Success: false, Message: "Unsupported pod network, please use 'cilium', 'flannel' or 'weave'"}); err != nil {
			return err
		}
		return nil
	}

	found, _ = exists (kured_yaml)
	if found != true {
		if err := stream.Send(&pb.StatusReply{Success: false, Message: "kured-k8s-yaml is not installed!"}); err != nil {
			return err
		}
		return nil
	}


	success, message := tools.ExecuteCmd("systemctl", "enable", "--now", "crio")
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
			return err
		}
		return nil
	}
	success, message = tools.ExecuteCmd("systemctl", "enable", "--now", "kubelet")
	if success != true {
		tools.ExecuteCmd("systemctl", "disable", "--now", "crio")
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
			return err
		}
		return nil
	}

	// build kubeadm call
	kubeadm_args := []string{"init"}

	if len(in.AdvAddr) > 0 {
		kubeadm_args = append(kubeadm_args, "--apiserver-advertise-address=" + in.AdvAddr)
	}

	if strings.EqualFold(arg_pod_network, "flannel") {
		kubeadm_args = append(kubeadm_args, "--pod-network-cidr=10.244.0.0/16")
	}

	if len (in.KubernetesVersion) > 0 {
		kubeadm_args = append(kubeadm_args, "--kubernetes-version=" + in.KubernetesVersion)
		update_cfg ("control-plane.conf", "version", in.KubernetesVersion)
	} else {
		// No version given. Try to use kubeadm RPM version number.
		success, message := tools.ExecuteCmd("rpm", "-q", "--qf", "'%{VERSION}'",  "kubernetes-kubeadm")
		if success == true {
			kubernetes_version := strings.Replace(message, "'", "", -1)
			kubeadm_args = append(kubeadm_args, "--kubernetes-version=" + kubernetes_version)
			update_cfg ("control-plane.conf", "version", kubernetes_version)
		}
	}

	if err := stream.Send(&pb.StatusReply{Success: true, Message: "Initialize Kubernetes control-plane"}); err != nil {
		return err
	}
	log.Infof ("Calling kubeadm '%v'", kubeadm_args)
	success, message = tools.ExecuteCmd("kubeadm", kubeadm_args...)
	if success != true {
		ResetMaster()
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
			return err
		}
		return nil
	}

	if strings.EqualFold(arg_pod_network, "weave") {
		// Setting up weave
		if err := stream.Send(&pb.StatusReply{Success: true, Message: "Deploy weave"}); err != nil {
			return err
		}
		success, message = deployment.DeployFile(weave_yaml)
		if success != true {
			ResetMaster()
			if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
				return err
			}
			return nil
		}
	} else if strings.EqualFold(arg_pod_network, "flannel") {
		// Setting up flannel
		if err := stream.Send(&pb.StatusReply{Success: true, Message: "Deploy flannel"}); err != nil {
			return err
		}
		success, message = deployment.DeployFile(flannel_yaml)
		if success != true {
			ResetMaster()
			if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
				return err
			}
			return nil
		}
	} else if strings.EqualFold(arg_pod_network, "cilium") {
		// Setting up cilium
		if err := stream.Send(&pb.StatusReply{Success: true, Message: "Deploy cilium"}); err != nil {
			return err
		}
		success, message = deployment.DeployFile(cilium_yaml)
		if success != true {
			ResetMaster()
			if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
				return err
			}
			return nil
		}
	}

	// Setting up kured
	if err := stream.Send(&pb.StatusReply{Success: true, Message: "Deploy Kubernetes Reboot Daemon (kured)"}); err != nil {
		return err
	}
	success, message = deployment.DeployFile(kured_yaml)
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
