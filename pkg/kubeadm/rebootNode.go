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

func RebootNode(nodeName string) (bool, string) {

	// salt host names are not identical with kubernetes node name.
	hostname, err := tools.GetNodeName(nodeName)
	if err != nil {
		return false, err.Error()
	}

	success, message := tools.DrainNode(hostname, "")
	if success != true {
		return success, message
	}

	success, message = tools.ExecuteCmd("salt", nodeName, "system.reboot")
	if success != true {
		return success, message
	}

	return true, ""
}
