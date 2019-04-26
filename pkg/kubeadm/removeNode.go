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
)


func RemoveNode(in *pb.RemoveNodeRequest, stream pb.Kubeadm_RemoveNodeServer) error {
	// XXX in.NodeNames could be a list of Nodes ...
	// salt host names are not identical with kubernetes node name.
	hostname, herr := GetNodeName(in.NodeNames)
	if herr != nil {
		if err := stream.Send(&pb.StatusReply{Success: false, Message: herr.Error()}); err != nil {
                        return err
                }
                return nil
	}

	if err := stream.Send(&pb.StatusReply{Success: true, Message: "Draining node " + hostname + "..."}); err != nil {
		return err
	}

	success, message := ExecuteCmd("kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
		"drain",  hostname, "--delete-local-data",  "--force",  "--ignore-daemonsets")
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
                        return err
                }
                return nil
	}

	if err := stream.Send(&pb.StatusReply{Success: true, Message: "Removing node " + hostname + "from Kubernetes"}); err != nil {
		return err
	}
	success, message = ExecuteCmd("kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
		"delete",  "node",  hostname)
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
                        return err
                }
                return nil
	}

	if err := stream.Send(&pb.StatusReply{Success: true, Message: "Cleanup node " + hostname + "..."}); err != nil {
		return err
	}
	success, message = ExecuteCmd("salt", in.NodeNames, "cmd.run",  "kubeadm reset --force")
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: success, Message: message}); err != nil {
                        return err
                }
                return nil
	}
	// Try some system cleanup, ignore if fails
	ExecuteCmd("salt", in.NodeNames, "cmd.run", "sed -i -e 's|^REBOOT_METHOD=kured|REBOOT_METHOD=auto|g' /etc/transactional-update.conf")
	ExecuteCmd("salt", in.NodeNames, "grains.delkey",  "kubicd")
	success, message = ExecuteCmd("salt", in.NodeNames, "cmd.run",  "\"iptables -t nat -F && iptables -t mangle -F && iptables -X\"")
	if err := stream.Send(&pb.StatusReply{Success: true, Message: "Warning: removal of iptables failed: "+message}); err != nil {
		return err
	}
	success, message = ExecuteCmd("salt", in.NodeNames, "cmd.run",  "\"ip link delete cni0;  ip link delete flannel.1\"")
	if err := stream.Send(&pb.StatusReply{Success: true, Message: "Warning: removal of network interfaces failed: "+message}); err != nil {
		return err
	}
	ExecuteCmd("salt", in.NodeNames, "service.disable",  "kubelet")
	ExecuteCmd("salt", in.NodeNames, "service.stop",  "kubelet")
	ExecuteCmd("salt", in.NodeNames, "service.disable",  "crio")
	ExecuteCmd("salt", in.NodeNames, "service.stop",  "crio")
	return nil
}
