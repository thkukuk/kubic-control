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

func AddNode(nodeNames string) (bool, string) {
	arg_socket := "--cri-socket=/run/crio/crio.sock"

	// XXX Store join command for 23 hours 30 minutes and re-use it.
	success, joincmd := ExecuteCmd("kubeadm", "token", "create", "--print-join-command")
	if success != true {
		return success, joincmd
	}

	joincmd = strings.TrimSuffix(joincmd, "\n")

	var message string
	// Differentiate between 'name1,name2' and 'name[1,2]'
	if strings.Index(nodeNames, ",") >= 0 && strings.Index(nodeNames, "[") == -1 {
		success, message = ExecuteCmd("salt", "-L", nodeNames, "cmd.run", "\"" + joincmd + " " + arg_socket + "\"")
		if success != true {
			return success, message
		}
	} else {
		success, message = ExecuteCmd("salt",  nodeNames, "cmd.run",  "\"" + joincmd + " " + arg_socket + "\"")
		if success != true {
			return success, message
		}
	}

	return true, ""
}
