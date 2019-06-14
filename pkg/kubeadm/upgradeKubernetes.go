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
	"strings"

	pb "github.com/thkukuk/kubic-control/api"
	"github.com/thkukuk/kubic-control/pkg/tools"
)

func UpgradeKubernetes(in *pb.Empty, stream pb.Kubeadm_UpgradeKubernetesServer) error {
	// find out our kubeadm version and use that to upgrade to this version
	success, message := tools.ExecuteCmd("rpm", "-q", "--qf", "'%{VERSION}'",  "kubernetes-kubeadm")
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
                        return err
                }
                return nil
	}
	kubernetes_version := strings.Replace(message, "'", "", -1)

	// Check if kuberadm and kubelet is new enough on all nodes
	// salt '*' --out=yaml pkg.version kubernetes-kubeadm kubernetes-kubelet

	if err := stream.Send(&pb.StatusReply{Success: true, Message: "Validate whether the cluster is upgradeable..."}); err != nil {
		return err
	}
	success, message = tools.ExecuteCmd("kubeadm",  "upgrade", "plan", kubernetes_version)
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
                        return err
                }
                return nil
	}

	if err := stream.Send(&pb.StatusReply{Success: true, Message: "Upgrade the control plane..."}); err != nil {
		return err
	}
	success, message = tools.ExecuteCmd("kubeadm",  "upgrade", "apply", "v"+kubernetes_version, "--yes")
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
                        return err
                }
                return nil
	}

	// Get list of all worker nodes:
	success, message = tools.ExecuteCmd("salt", "-G", "kubicd:kubic-worker-node", "grains.get",  "kubic-worker-node")
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
                        return err
                }
                return nil
	}
	message = strings.TrimSuffix(message, "\n")
	nodelist := strings.Split (strings.Replace(message, ":", "", -1), "\n")

	var failedNodes = ""
	for i := range nodelist {
		if err := stream.Send(&pb.StatusReply{Success: true, Message: "Upgrade "+nodelist[i]+"..."}); err != nil {
			return err
		}
		hostname, err := GetNodeName(nodelist[i])
		if err != nil {
			failedNodes = failedNodes+nodelist[i]+" (uncordon), "
		} else {
			// if draining fails, ignore
			tools.DrainNode(hostname, "")

			success,message = tools.ExecuteCmd("salt", nodelist[i], "cmd.run",
				"\"kubeadm upgrade node config --kubelet-version " + kubernetes_version + "\"")
			if success != true {
				failedNodes = failedNodes+nodelist[i]+" (kubeadm), "
			} else {
				success,message = tools.ExecuteCmd("kubectl", "--kubeconfig=/etc/kubernetes/admin.conf", "uncordon",  hostname)
				if success != true {
					failedNodes = failedNodes+nodelist[i]+" (uncordon), "
				}
			}
		}
	}

	if len(failedNodes) > 0 {
		// XXX remove ", " Suffix
		if err := stream.Send(&pb.StatusReply{Success: false, Message: failedNodes}); err != nil {
			return err
		}
	}

	if err := stream.Send(&pb.StatusReply{Success: true, Message: "Kubernetes cluster was successfully upgraded to version " + kubernetes_version}); err != nil {
		return err
	}
	return nil
}
