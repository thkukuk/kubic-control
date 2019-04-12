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
)

func RemoveNode(nodeName string) (bool, string) {

	// salt host names are not identical with kubernetes node name.
	// Output of hostname should be identical to node name
	success, message := ExecuteCmd("salt",  nodeName, "cmd.run",  "hostname")
	if success != true {
		return success, message
	}
	hostname := strings.Replace(message, "\n","",-1)
	i := strings.Index(hostname,":")+1
	hostname = hostname[i:]
	hostname = strings.TrimSpace(hostname)

	success, message = ExecuteCmd("kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
		"drain",  hostname, "--delete-local-data",  "--force",  "--ignore-daemonsets")
	if success != true {
		return success, message
	}
	success, message = ExecuteCmd("kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
		"delete",  "node",  hostname)
	if success != true {
		return success, message
	}

	success, message = ExecuteCmd("salt",  nodeName, "cmd.run",  "\"kubeadm reset --force\"")
	if success != true {
		return success, message
	}
	// Try some system cleanup, ignore if fails
	ExecuteCmd("salt",  nodeName, "cmd.run",  "\"iptables -t nat -F && iptables -t mangle -F && iptables -X\"")
	ExecuteCmd("salt",  nodeName, "cmd.run",  "\"ip link delete cni0;  ip link delete flannel.1\"")
	ExecuteCmd("salt",  nodeName, "cmd.run",  "\"systemctl disable --now crio\"")
	ExecuteCmd("salt",  nodeName, "cmd.run",  "\"systemctl disable --now kubelet\"")
	return true, ""
}
