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

func UpgradeKubernetes(Version string) (bool, string) {
	// find out our kubeadm version and use that to upgrade to this version
	success, message := ExecuteCmd("rpm", "-q", "--qf", "'%{VERSION}'",  "kubernetes-kubeadm")
	if success != true {
		return success, message
	}
	kubernetes_version := strings.Replace(message, "'", "", -1)

	// Check if kuberadm and kubelet is new enough on all nodes
	// salt '*' --out=yaml pkg.version kubernetes-kubeadm kubernetes-kubelet

	success, message = ExecuteCmd("kubeadm",  "upgrade", "plan", kubernetes_version)
	if success != true {
		return success, message
	}

	success, message = ExecuteCmd("kubeadm",  "upgrade", "apply", "v"+kubernetes_version, "--yes")
	if success != true {
		return success, message
	}

	// Get list of all worker nodes:
	success, message = ExecuteCmd("salt", "-G", "kubicd:kubic-worker-node", "grains.get",  "kubic-worker-node")
	if success != true {
		return success, message
	}
	message = strings.TrimSuffix(message, "\n")
	nodelist := strings.Split (strings.Replace(message, ":", "", -1), "\n")

	var failedNodes = ""
	for i := range nodelist {
		hostname := GetNodeName(nodelist[i])
		// if draining fails, ignore
		ExecuteCmd("kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
			"drain",  hostname,  "--force",  "--ignore-daemonsets")

		success,message = ExecuteCmd("salt", nodelist[i], "cmd.run",
			"\"kubeadm upgrade node config --kubelet-version " + kubernetes_version + "\"")
		if success != true {
			failedNodes = failedNodes+nodelist[i]+" (kubeadm), "
		} else {
			success,message = ExecuteCmd("kubectl", "--kubeconfig=/etc/kubernetes/admin.conf", "uncordon",  hostname)
			if success != true {
				failedNodes = failedNodes+nodelist[i]+" (uncordon), "
			}
		}
	}

	if len(failedNodes) > 0 {
		return false, failedNodes
	}

	return true, ""
}
