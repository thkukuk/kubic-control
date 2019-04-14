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

func RebootNode(nodeName string) (bool, string) {

	// salt host names are not identical with kubernetes node name.
	hostname := GetNodeName(nodeName)

	success, message := ExecuteCmd("kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
		"drain",  hostname, "--force",  "--ignore-daemonsets")
	if success != true {
		return success, message
	}

	success, message = ExecuteCmd("salt",  nodeName, "sytem.reboot")
	if success != true {
		return success, message
	}

	return true, ""
}
