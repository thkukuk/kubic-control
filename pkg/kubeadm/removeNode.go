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

func RemoveNode(nodeName string) (bool, string) {

	// salt host names are not identical with kubernetes node name.
	hostname, err := GetNodeName(nodeName)
	if err != nil {
		return false, err.Error()
	}

	success, message := ExecuteCmd("kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
		"drain",  hostname, "--delete-local-data",  "--force",  "--ignore-daemonsets")
	if success != true {
		return success, message
	}
	success, message = ExecuteCmd("kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
		"delete",  "node",  hostname)
	if success != true {
		return success, message
	}

	success, message = ExecuteCmd("salt", nodeName, "cmd.run",  "kubeadm reset --force")
	if success != true {
		return success, message
	}
	// Try some system cleanup, ignore if fails
	ExecuteCmd("salt", nodeName, "cmd.run", "sed -i -e 's|^REBOOT_METHOD=kured|REBOOT_METHOD=auto|g' /etc/transactional-update.conf")
	ExecuteCmd("salt", nodeName, "grains.delkey",  "kubicd")
	ExecuteCmd("salt", nodeName, "cmd.run",  "\"iptables -t nat -F && iptables -t mangle -F && iptables -X\"")
	ExecuteCmd("salt", nodeName, "cmd.run",  "\"ip link delete cni0;  ip link delete flannel.1\"")
	ExecuteCmd("salt", nodeName, "service.disable",  "kubelet")
	ExecuteCmd("salt", nodeName, "service.stop",  "kubelet")
	ExecuteCmd("salt", nodeName, "service.disable",  "crio")
	ExecuteCmd("salt", nodeName, "service.stop",  "crio")
	return true, ""
}
