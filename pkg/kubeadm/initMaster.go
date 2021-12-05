// Copyright 2019, 2020 Thorsten Kukuk
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
	"errors"
	"os"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	pb "github.com/thkukuk/kubic-control/api"
	"github.com/thkukuk/kubic-control/pkg/deployment"
	"github.com/thkukuk/kubic-control/pkg/tools"
	"gopkg.in/ini.v1"
)

const (
	flannel_yaml     = "/usr/share/k8s-yaml/flannel/kube-flannel.yaml"
	weave_yaml       = "/usr/share/k8s-yaml/weave/weave.yaml"
	kured_yaml       = "/usr/share/k8s-yaml/kured/kured.yaml"
)

// update data in /var/lib/kubic-control
func update_cfg(file string, key string, value string) error {
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

func executeCmdSalt(salt string, command string, arg ...string) (bool, string) {
	if len(salt) > 0 {
		return tools.ExecuteCmd("salt", salt, "cmd.run", command+" "+strings.Join(arg[:], " "))
	} else {
		return tools.ExecuteCmd(command, arg...)
	}
}

// exists returns whether the given file or directory exists
func exists(path string, salt string) (bool, error) {
	if len(salt) > 0 {
		success, message := tools.ExecuteCmd("salt", "--out=txt", salt, "file.access", path, "f")
		if success != true {
			return false, errors.New(message)
		}
		if strings.HasSuffix(message, ": True") {
			return true, nil
		} else {
			return false, nil
		}
	} else {
		_, err := os.Stat(path)
		if err == nil {
			return true, nil
		}
		if os.IsNotExist(err) {
			return false, nil
		}
		return true, err
	}
}

func InitMaster(in *pb.InitRequest, stream pb.Kubeadm_InitMasterServer) error {
	arg_pod_network := in.PodNetworking
	arg_salt := in.FirstMaster

	found, _ := exists("/etc/kubernetes/manifests/kube-apiserver.yaml", arg_salt)
	if found == true {
		if err := stream.Send(&pb.StatusReply{Success: false, Message: "Seems like a kubernetes control-plane is already running. If not, please use \"kubeadm reset\" to clean up the system."}); err != nil {
			return err
		}
		return nil
	}
	found, _ = exists("/etc/kubernetes/manifests/kube-scheduler.yaml", arg_salt)
	if found == true {
		if err := stream.Send(&pb.StatusReply{Success: false, Message: "Seems like a kubernetes control-plane is already running. If not, please use \"kubeadm reset\" to clean up the system"}); err != nil {
			return err
		}
		return nil
	}
	found, _ = exists("/etc/kubernetes/manifests/etcd.yaml", arg_salt)
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
		found, _ = exists(weave_yaml, "")
		if found != true {
			if err := stream.Send(&pb.StatusReply{Success: false, Message: "weave-k8s-yaml is not installed!"}); err != nil {
				return err
			}
			return nil
		}
	} else if strings.EqualFold(arg_pod_network, "flannel") {
		found, _ = exists(flannel_yaml, "")
		if found != true {
			if err := stream.Send(&pb.StatusReply{Success: false, Message: "flannel-k8s-yaml is not installed!"}); err != nil {
				return err
			}
			return nil
		}
	} else if !strings.EqualFold(arg_pod_network, "none") {
		if err := stream.Send(&pb.StatusReply{Success: false, Message: "Unsupported pod network, please use 'flannel', 'weave' or 'none'"}); err != nil {
			return err
		}
		return nil
	}

	found, _ = exists(kured_yaml, "")
	if found != true {
		if err := stream.Send(&pb.StatusReply{Success: false, Message: "kured-k8s-yaml is not installed!"}); err != nil {
			return err
		}
		return nil
	}

	success, message := executeCmdSalt(arg_salt, "systemctl", "enable", "--now", "crio")
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
			return err
		}
		return nil
	}
	success, message = executeCmdSalt(arg_salt, "systemctl", "enable", "--now", "kubelet")
	if success != true {
		executeCmdSalt(arg_salt, "systemctl", "disable", "--now", "crio")
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
			return err
		}
		return nil
	}

	if len(in.MultiMaster) > 0 {
		message = "Setting up multi-master kubernetes node (reacheable as '" + in.MultiMaster + "') with " + arg_pod_network
		if err := stream.Send(&pb.StatusReply{Success: true, Message: message}); err != nil {
			return err
		}
		if len(in.Haproxy) > 0 {
			message = "Configure haproxy on node " + in.Haproxy
			if err := stream.Send(&pb.StatusReply{Success: true, Message: message}); err != nil {
				return err
			}
			hostname, err := os.Hostname()
			if err != nil {
				if err2 := stream.Send(&pb.StatusReply{Success: false,
					Message: "Could not get hostname: " + err.Error() +
						"\nPlease setup your haproxy manually before continuing"}); err2 != nil {
					return err2
				}
				return nil
			}
			success, message = tools.ExecuteCmd("salt", in.Haproxy, "cmd.run",
				"\"haproxycfg init --force "+in.MultiMaster+" "+hostname+"\"")
			if success != true {
				if err := stream.Send(&pb.StatusReply{Success: false, Message: message}); err != nil {
					return err
				}
				return nil
			}
		}
	} else {
		message = "Setting up single-master kubernetes node with " + arg_pod_network
		if err := stream.Send(&pb.StatusReply{Success: true, Message: message}); err != nil {
			return err
		}
	}

	// build kubeadm call
	kubeadm_args := []string{"init"}

	if strings.EqualFold(arg_pod_network, "flannel") {
		kubeadm_args = append(kubeadm_args, "--pod-network-cidr=10.244.0.0/16")
	}

	if len(in.Stage) > 0 {
		if strings.EqualFold(in.Stage, "devel") {
			if runtime.GOARCH == "amd64" {
				kubeadm_args = append(kubeadm_args, "--image-repository=registry.opensuse.org/devel/kubic/containers/container/kubic")
			} else if runtime.GOARCH == "arm64" {
				kubeadm_args = append(kubeadm_args, "--image-repository=registry.opensuse.org/devel/kubic/containers/container_arm/kubic")
			} else {
				message = "Unknown architecture '" + runtime.GOARCH + "', no devel project known, using standard one"
				if err := stream.Send(&pb.StatusReply{Success: true, Message: message}); err != nil {
					return err
				}
			}
		} else if !strings.EqualFold(in.Stage, "official") {
			/* Ugly hack, we will use the argument as pointer to a registry */
			kubeadm_args = append(kubeadm_args, "--image-repository="+in.Stage)
		}
	}

	kubernetes_version := ""
	if len(in.KubernetesVersion) > 0 {
		kubernetes_version = in.KubernetesVersion
	} else {
		success, message := tools.GetKubeadmVersion(arg_salt)
		if success != true {
			if err := stream.Send(&pb.StatusReply{Success: false, Message: message}); err != nil {
				return err
			}
			return nil
		}
		kubernetes_version = message
	}
	update_cfg("control-plane.conf", "version", kubernetes_version)
	update_cfg("control-plane.conf", "master", arg_salt)

	if len(in.MultiMaster) > 0 {
		os.MkdirAll("/var/lib/kubic-control/multi-master", os.ModePerm)

		f, err := os.Create("/var/lib/kubic-control/multi-master/kubeadm-config.yaml")
		if err != nil {
			ResetMaster()
			if err := stream.Send(&pb.StatusReply{Success: false, Message: err.Error()}); err != nil {
				return err
			}
			return nil
		}
		defer f.Close()

		_, err = f.WriteString("apiVersion: kubeadm.k8s.io/v1beta2\nkind: ClusterConfiguration\nkubernetesVersion: " + kubernetes_version + "\ncontrolPlaneEndpoint: \"" + in.MultiMaster + ":6443\"\n")
		if err != nil {
			ResetMaster()
			if err := stream.Send(&pb.StatusReply{Success: false, Message: err.Error()}); err != nil {
				return err
			}
			return nil
		}

		if len(in.ApiserverCertExtraSans) > 0 || len(in.AdvAddr) > 0 {
			_, err = f.WriteString("apiServer:\n")
			if err != nil {
				ResetMaster()
				if err := stream.Send(&pb.StatusReply{Success: false, Message: err.Error()}); err != nil {
					return err
				}
				return nil
			}

			if len(in.ApiserverCertExtraSans) > 0 {
				_, err = f.WriteString("  certSANs:\n    - " + in.ApiserverCertExtraSans + "\n")
				if err != nil {
					ResetMaster()
					if err := stream.Send(&pb.StatusReply{Success: false, Message: err.Error()}); err != nil {
						return err
					}
					return nil
				}
			}

			if len(in.AdvAddr) > 0 {
				_, err = f.WriteString("  extraArgs:\n    advertise-address: " + in.AdvAddr + "\n")
				if err != nil {
					ResetMaster()
					if err := stream.Send(&pb.StatusReply{Success: false, Message: err.Error()}); err != nil {
						return err
					}
					return nil
				}
			}
		}
		f.Close()

		update_cfg("control-plane.conf", "MultiMaster", "True")
		update_cfg("control-plane.conf", "loadbalancer_dns", in.MultiMaster)
		if len(in.Haproxy) > 0 {
			update_cfg("control-plane.conf", "loadbalancer_salt", in.Haproxy)
		}

		kubeadm_args = append(kubeadm_args,
			"--config=/var/lib/kubic-control/multi-master/kubeadm-config.yaml")
		// No need to upload certs, we have to do it anyways if we add a new
		// master node.
		// kubeadm_args = append(kubeadm_args, "--upload-certs")
	} else {
		// kubeadm does not really like mixing config files and arguments, only use
		// --kubernetes-version if we don't use a config file.
		kubeadm_args = append(kubeadm_args, "--kubernetes-version="+kubernetes_version)

		if len(in.AdvAddr) > 0 {
			kubeadm_args = append(kubeadm_args, "--apiserver-advertise-address="+in.AdvAddr)
		}

		if len(in.ApiserverCertExtraSans) > 0 {
			kubeadm_args = append(kubeadm_args, "--apiserver-cert-extra-sans="+in.ApiserverCertExtraSans)
		}
	}

	if err := stream.Send(&pb.StatusReply{Success: true, Message: "Initialize Kubernetes control-plane"}); err != nil {
		return err
	}
	log.Infof("Calling kubeadm '%v'", kubeadm_args)
	success, message = executeCmdSalt(arg_salt, "kubeadm", kubeadm_args...)
	if success != true {
		ResetMaster()
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
			return err
		}
		return nil
	}

	if len(arg_salt) > 0 {
		// Get kubernetes/admin.conf for kubectl calls
		tools.ExecuteCmd("mkdir", "/etc/kubernetes")
		log.Infof("Download /etc/kubernetes/admin.conf")
		success, message = tools.ExecuteCmd("salt", "--out=newline_values_only",
			"--out-file=/etc/kubernetes/admin.conf", arg_salt,
			"cmd.run", "cat /etc/kubernetes/admin.conf")
		if success != true {
			ResetMaster()
			if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
				return err
			}
			return nil
		}
		os.Chmod("/etc/kubernetes/admin.conf", 0600) // XXX error handling
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
	} else if strings.EqualFold(arg_pod_network, "none") {
		if err := stream.Send(&pb.StatusReply{Success: true, Message: "No CNI will be deployed"}); err != nil {
			return err
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
	if len(arg_salt) > 0 {
		success, message = tools.ExecuteCmd("salt", arg_salt, "grains.append", "kubicd", "kubic-master-node")
		if success != true {
			if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
				return err
			}
		}
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

	if len(in.MultiMaster) > 0 {
		if err := stream.Send(&pb.StatusReply{Success: true, Message: "First Kubernetes master succesfully setup."}); err != nil {
			return err
		}
		if err := stream.Send(&pb.StatusReply{Success: true, Message: "Please add at minimum two further master nodes!"}); err != nil {
			return err
		}
	} else {
		if err := stream.Send(&pb.StatusReply{Success: true, Message: "Kubernetes master was succesfully setup."}); err != nil {
			return err
		}
	}
	return nil
}
