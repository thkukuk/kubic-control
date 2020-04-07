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
	"os"
	"strings"

	pb "github.com/thkukuk/kubic-control/api"
	"github.com/thkukuk/kubic-control/pkg/tools"
	"github.com/thkukuk/kubic-control/pkg/deployment"
)

func uncordon(stream pb.Kubeadm_UpgradeKubernetesServer, hostname string) error {
	// uncordon
	success, message := tools.ExecuteCmd("kubectl", "--kubeconfig=/etc/kubernetes/admin.conf", "uncordon",  hostname)
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
                        return err
                }
	}
	return nil
}

func upgradeFirstMaster(in *pb.UpgradeRequest, stream pb.Kubeadm_UpgradeKubernetesServer, kubernetes_version string) error {
	hostname, err := os.Hostname();
	if err != nil {
		if err2 := stream.Send(&pb.StatusReply{Success: false,
			Message: "Could not get hostname: " + err.Error()}); err2 != nil {
				return err2
			}
		return nil
	}
	if err := stream.Send(&pb.StatusReply{Success: true, Message: "Validate whether the cluster is upgradeable..."}); err != nil {
		return err
	}
	success, message := tools.ExecuteCmd("kubeadm",  "upgrade", "plan", kubernetes_version)
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
                        return err
                }
                return nil
	}

	if err := stream.Send(&pb.StatusReply{Success: true, Message: "Drain first control plane master..."}); err != nil {
		return err
	}
	// if draining fails, ignore
	tools.DrainNode(hostname, "")

	if err := stream.Send(&pb.StatusReply{Success: true, Message: "Upgrade the control plane..."}); err != nil {
		uncordon(stream, hostname)
		return err
	}
	success, message = tools.ExecuteCmd("kubeadm",  "upgrade", "apply", kubernetes_version, "--yes")
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
			uncordon(stream, hostname)
                        return err
                }
		uncordon(stream, hostname)
                return nil
	}
	// Update kubelet
	success, message = tools.ExecuteCmd("sed", "-i",  "'s/KUBELET_VER=1.17/KUBELET_VER=1.18/'", "/etc/sysconfig/kubelet")
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
			uncordon(stream, hostname)
                        return err
                }
		uncordon(stream, hostname)
                return nil
	}
	success, message = tools.ExecuteCmd("systemctl", "restart", "kubelet")
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
			uncordon(stream, hostname)
                        return err
                }
		uncordon(stream, hostname)
                return nil
	}
	return uncordon(stream, hostname)
}

func upgradeNodes(in *pb.UpgradeRequest,
	stream pb.Kubeadm_UpgradeKubernetesServer,
	role string, kubernetes_version string) (string, error) {
	// Get list of all role nodes:
	success, message, nodelist := tools.GetListOfNodes(role)
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
                        return "", err
                }
                return "", nil
	}

	var failedNodes = ""
	for i := range nodelist {
		if err := stream.Send(&pb.StatusReply{Success: true, Message: "Upgrade "+nodelist[i]+"..."}); err != nil {
			return "", err
		}
		hostname, err := GetNodeName(nodelist[i])
		if err != nil {
			failedNodes = failedNodes+nodelist[i] + "(determine hostname), "
		} else {
			// if draining fails, ignore
			tools.DrainNode(hostname, "")

			success,message = tools.ExecuteCmd("salt", nodelist[i], "cmd.run",
				"\"kubeadm upgrade node --kubelet-version " + kubernetes_version + "\"")
			if success != true {
				failedNodes = failedNodes+nodelist[i]+" (kubeadm), "
			} else {
				// Update kubelet
				success, message = tools.ExecuteCmd("salt", nodelist[i], "cmd.run",
					"\"sed -i 's/KUBELET_VER=1.17/KUBELET_VER=1.18/' /etc/sysconfig/kubelet\"")
				if success != true {
					failedNodes = failedNodes+nodelist[i]+" (kubelet_ver), "
				} else {
					success, message = tools.ExecuteCmd("salt", nodelist[i], "service.restart", "kubelet")
					if success != true {
						failedNodes = failedNodes+nodelist[i]+" (kubelet), "
					}
				}
			}
			// uncordon, most likely node will still work, else we can run out of nodes
			success,message = tools.ExecuteCmd("kubectl", "--kubeconfig=/etc/kubernetes/admin.conf", "uncordon",  hostname)
			if success != true {
				failedNodes = failedNodes+nodelist[i]+" (uncordon), "
			}
		}
	}
	return failedNodes, nil
}

func UpgradeKubernetes(in *pb.UpgradeRequest, stream pb.Kubeadm_UpgradeKubernetesServer) error {

	multiMaster := Read_Cfg("control-plane.conf", "MultiMaster")

	kubernetes_version := ""
        if len (in.KubernetesVersion) > 0 {
                kubernetes_version = in.KubernetesVersion
        } else {
		success, message := tools.GetKubeadmVersion("") // XXX Upgrade needs to support remote master
                if success != true {
                        if err := stream.Send(&pb.StatusReply{Success: false, Message: message}); err != nil {
                                return err
                        }
                        return nil
                }
		kubernetes_version = message
        }

	// XXX Check if kuberadm is new enough on all nodes
	// salt '*' --out=txt pkg.version kubernetes-kubeadm

	if err := upgradeFirstMaster(in, stream, kubernetes_version); err != nil {
		return err
	}
	var failedMaster string
	if strings.EqualFold(multiMaster, "True") {
		var err error
		if failedMaster, err = upgradeNodes(in, stream, "master", kubernetes_version); err != nil {
			return err
		}
	}
	var failedWorker string
	{
		var err error
		if failedWorker, err = upgradeNodes(in, stream, "role", kubernetes_version); err != nil {
			return err
		}
	}


	// Update pod network, kured and other pods we are running:
	success, message := deployment.UpdateAll(false)
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
                        return err
                }
	}

	if len(failedMaster) > 0 || len(failedWorker) > 0 {
		if len(failedMaster) > 0 {
			if err := stream.Send(&pb.StatusReply{Success: false, Message: "Upgrade of some master nodes failed: " + strings.TrimSuffix(failedMaster, ", ")}); err != nil {
				return err
			}
		}
		if len(failedWorker) > 0 {
			if err := stream.Send(&pb.StatusReply{Success: false, Message: "Upgrade of some Nodes failed: " + strings.TrimSuffix(failedWorker, ", ")}); err != nil {
				return err
			}
		}
	} else {
		if err := stream.Send(&pb.StatusReply{Success: true, Message: "Kubernetes cluster was successfully upgraded to version " + kubernetes_version}); err != nil {
			return err
		}
	}
	return nil
}
