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
	pb "github.com/thkukuk/kubic-control/api"
	"github.com/thkukuk/kubic-control/pkg/tools"
	"github.com/thkukuk/kubic-control/pkg/deployment"
)

func UpgradeKubernetes(in *pb.UpgradeRequest, stream pb.Kubeadm_UpgradeKubernetesServer) error {

	kubernetes_version := ""
        if len (in.KubernetesVersion) > 0 {
                kubernetes_version = in.KubernetesVersion
        } else {
		success, message := tools.GetKubeadmVersion()
                if success != true {
                        if err := stream.Send(&pb.StatusReply{Success: false, Message: message}); err != nil {
                                return err
                        }
                        return nil
                }
                kubernetes_version = message
        }

	// XXX Check if kuberadm and kubelet is new enough on all nodes
	// salt '*' --out=txt pkg.version kubernetes-kubeadm kubernetes-kubelet

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

	if err := stream.Send(&pb.StatusReply{Success: true, Message: "Upgrade the control plane..."}); err != nil {
		return err
	}
	success, message = tools.ExecuteCmd("kubeadm",  "upgrade", "apply", kubernetes_version, "--yes")
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
                        return err
                }
                return nil
	}

	// Get list of all worker nodes:
	success, message, nodelist := tools.GetListOfNodes()
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
                        return err
                }
                return nil
	}

	var failedNodes = ""
	for i := range nodelist {
		if err := stream.Send(&pb.StatusReply{Success: true, Message: "Upgrade "+nodelist[i]+"..."}); err != nil {
			return err
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
				success,message = tools.ExecuteCmd("kubectl", "--kubeconfig=/etc/kubernetes/admin.conf", "uncordon",  hostname)
				if success != true {
					failedNodes = failedNodes+nodelist[i]+" (uncordon), "
				}
			}
		}
	}

	// Update pod network, kured and other pods we are running:
	success, message = deployment.UpdateAll(false)
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
                        return err
                }
	}

	if len(failedNodes) > 0 {
		// XXX remove ", " Suffix
		if err := stream.Send(&pb.StatusReply{Success: false, Message: "Upgrade of some Nodes failed: " + failedNodes}); err != nil {
			return err
		}
	} else {
		if err := stream.Send(&pb.StatusReply{Success: true, Message: "Kubernetes cluster was successfully upgraded to version " + kubernetes_version}); err != nil {
			return err
		}
	}
	return nil
}
