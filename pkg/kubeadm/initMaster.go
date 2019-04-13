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

func InitMaster(podNetwork string, kubernetesVersion string) (bool, string) {
	arg_socket := "--cri-socket=/run/crio/crio.sock"
	arg_pod_network_cidr := ""
	arg_kubernetes_version := ""

	success, message := ExecuteCmd("systemctl", "enable", "--now", "crio")
	if success != true {
		return success, message
	}
	success, message = ExecuteCmd("systemctl", "enable", "--now", "kubelet")
	if success != true {
		ExecuteCmd("systemctl", "disable", "--now", "crio")
		return success, message
	}

	if (strings.EqualFold(podNetwork, "flannel")) {
		arg_pod_network_cidr = "--pod-network-cidr=10.244.0.0/16"
	}
	if len (kubernetesVersion) > 0 {
		arg_kubernetes_version = "--kubernetes-version=" + kubernetesVersion
	} else {
		// No version given. Try to use kubeadm RPM version number.
		success, message := ExecuteCmd("rpm", "-q", "--qf", "'%{VERSION}'",  "kubernetes-kubeadm")
		if success == true {
			arg_kubernetes_version = message
		}
	}

	success, message = ExecuteCmd("kubeadm", "init", arg_socket,
		arg_pod_network_cidr, arg_kubernetes_version)
	if success != true {
		ResetMaster()
		return success, message
	}

	// Setting up flannel
	success, message = ExecuteCmd("kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",  "apply", "-f", "https://raw.githubusercontent.com/coreos/flannel/bc79dd1505b0c8681ece4de4c0d86c5cd2643275/Documentation/kube-flannel.yml")
	if success != true {
		ResetMaster()
		return success, message
	}

	// Setting up kured
	success, message = ExecuteCmd("kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",  "apply", "-f", "/usr/share/k8s-yaml/kured/kured.yaml")
	if success != true {
		ResetMaster()
		return success, message
	}

	return true, ""
}
