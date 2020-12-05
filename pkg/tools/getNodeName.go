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
	"errors"
	"strings"
)

func GetNodeName(target string) (string, error) {

	// salt host names are not identical with kubernetes node name.
	// Output of hostname should be identical to node name
	success, message := ExecuteCmd("salt", target, "network.get_hostname")
	if success != true {
		return target, errors.New(message)
	}
	hostname := strings.Replace(message, "\n", "", -1)
	i := strings.Index(hostname, ":") + 1
	hostname = hostname[i:]
	return strings.TrimSpace(hostname), nil
}
