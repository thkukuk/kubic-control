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

func GetListOfNodes(role string) (bool, string, []string) {

	if len(role) == 0 {
		role = "worker"
	}

	// Get list of all nodes of this role
	success, message := ExecuteCmd("salt", "--module-executors='[direct_call]'", "-G", "kubicd:kubic-"+role+"-node", "grains.get", "kubic-"+role+"-node")
	if success != true {
		return success, message, nil
	}
	message = strings.TrimSuffix(message, "\n")
	nodelist := strings.Split(strings.Replace(message, ":", "", -1), "\n")

	return true, "", nodelist
}
