// Copyright 2019, 2020 Thorsten Kukuk
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

package tools

import (
	"strings"
)

func GetKubeadmVersion(salt string) (bool,string) {
	// find out our kubeadm version and use that to upgrade to this version
	var success bool
	var message string
	if len(salt) > 0 {
		success, message = ExecuteCmd("salt", salt, "cmd.run", "rpm -q --qf '%{VERSION}' kubernetes-kubeadm")
	} else {
		success, message = ExecuteCmd("rpm", "-q", "--qf", "'%{VERSION}'",  "kubernetes-kubeadm")
	}
	if success != true {
		return false, message
	}
	kubernetes_version := "v" + strings.Replace(message, "'", "", -1)

	return true, kubernetes_version
}
