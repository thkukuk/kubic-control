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

package yomi

import (
	"strings"
)

func Salt2PillarName(name string) string {
	// Replace all '.' with '_'
	temp := strings.Replace(name, ".", "_", -1)
	// Replace all spaces with '-'
	return strings.Replace(temp, " ", "-", -1)
}
