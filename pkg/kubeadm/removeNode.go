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

func RemoveNode(in *pb.RemoveNodeRequest, stream pb.Kubeadm_RemoveNodeServer) error {
	var nodelist []string

	// If we have a list of Nodes, try to find the right node names which have a kubic-worker-node grain.
	if strings.Index(in.NodeNames, ",") >= 0 || strings.Index(in.NodeNames, "[") >= 0 || strings.Compare(in.NodeNames, "*") == 0 {
		var success bool
		var message string

		if strings.Index(in.NodeNames, ",") >= 0 && strings.Index(in.NodeNames, "[")  == -1 {
			success, message = tools.ExecuteCmd("salt", "--out=txt", "-L", in.NodeNames, "grains.get", "kubicd")
		} else {
			success, message = tools.ExecuteCmd("salt", "--out=txt", in.NodeNames, "grains.get",  "kubicd")
		}
		if success != true {
			if err := stream.Send(&pb.StatusReply{Success: false, Message: message}); err != nil {
				return err
			}
			return nil
		}

		list := strings.Split (message, "\n")
		for _, entry := range list {
			if strings.Contains(entry, "'kubic-worker-node'") {
				list := strings.Split (entry, ":");
				nodelist = append (nodelist, list[0])
			}
		}
	} else {
		// only one node name to remove
		nodelist = append(nodelist,in.NodeNames)
	}

	for _, entry := range nodelist {
		stream.Send(&pb.StatusReply{Success: true, Message: entry});
	}

	// salt host names are not identical with kubernetes node name.
	var hostnames []string

	for _, entry := range nodelist {
		hostname, herr := GetNodeName(entry)
		if herr != nil {
			if err := stream.Send(&pb.StatusReply{Success: false, Message: herr.Error()}); err != nil {
				return err
			}
			return nil
		}
		hostnames = append (hostnames, hostname)
	}

	// loop over all hostnames, drain and delete them
	// XXX think about how to parallize
	for _, hostname := range hostnames {

		if err := stream.Send(&pb.StatusReply{Success: true, Message: "Draining node " + hostname + "..."}); err != nil {
			return err
		}

		success, message := tools.DrainNode(hostname, "")
		if success != true {
			if err := stream.Send(&pb.StatusReply{Success: true, Message: message + " (ignored)"}); err != nil {
				return err
			}
			// ignore error
		}

		if err := stream.Send(&pb.StatusReply{Success: true, Message: "Removing node " + hostname + " from Kubernetes"}); err != nil {
			return err
		}
		success, message = tools.ExecuteCmd("kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
			"delete",  "node",  hostname)
		if success != true {
			if err := stream.Send(&pb.StatusReply{Success: true, Message: message + " (ignored)"}); err != nil {
				return err
			}
			// ignore error
		}
	}


	salt_nodelist := strings.Join(nodelist, ",")
	if err := stream.Send(&pb.StatusReply{Success: true, Message: "Cleanup node(s) " + salt_nodelist + "..."}); err != nil {
		return err
	}
	success, message := tools.ExecuteCmd("salt", "-L", salt_nodelist, "cmd.run",  "kubeadm reset --force")
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: true, Message: message + " (ignored)"}); err != nil {
                        return err
                }
		// ignore error
	}
	// Try some system cleanup, ignore if fails
	tools.ExecuteCmd("salt", "-L", salt_nodelist, "cmd.run", "sed -i -e 's|^REBOOT_METHOD=kured|REBOOT_METHOD=auto|g' /etc/transactional-update.conf")
	tools.ExecuteCmd("salt", "-L", salt_nodelist, "grains.delkey",  "kubicd")
	success, message = tools.ExecuteCmd("salt", "-L", salt_nodelist, "cmd.run",  "\"iptables -F && iptables -t nat -F && iptables -t mangle -F && iptables -X\"")
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: true, Message: "Warning: removal of iptables failed."}); err != nil {
			return err
		}
	}
	tools.ExecuteCmd("salt", "-L", salt_nodelist, "cmd.run",  "\"ip link delete cni0;  ip link delete flannel.1; ip link delete cilium_vxlan\"")
	tools.ExecuteCmd("salt", "-L", salt_nodelist, "service.disable",  "kubelet")
	tools.ExecuteCmd("salt", "-L", salt_nodelist, "service.stop",  "kubelet")
	tools.ExecuteCmd("salt", "-L", salt_nodelist, "service.disable",  "crio")
	tools.ExecuteCmd("salt", "-L", salt_nodelist, "service.stop",  "crio")
	return nil
}
