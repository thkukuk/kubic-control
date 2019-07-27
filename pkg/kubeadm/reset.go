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
        "github.com/thkukuk/kubic-control/pkg/tools"
)

func ResetMaster() (bool, string) {
	arg_socket := "--cri-socket=/run/crio/crio.sock"

	success, message :=  tools.ExecuteCmd("kubeadm", "reset", "-f", arg_socket)

	// cleanup behind kubeadm
	tools.ExecuteCmd("rm", "-rf", "/var/lib/etcd/*")
	tools.ExecuteCmd("rm", "-rf", "/var/lib/cni/*")

	tools.ExecuteCmd("systemctl", "disable", "--now", "crio")
	tools.ExecuteCmd("systemctl", "disable", "--now", "kubelet")

	return success, message
}
