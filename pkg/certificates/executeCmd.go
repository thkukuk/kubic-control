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

package certificates

import (
	"os"
	"os/exec"
	"fmt"
	"bytes"
)

func ExecuteCmd(command string, arg ...string) (error,string) {
	var out bytes.Buffer
	var stderr bytes.Buffer

        cmd := exec.Command(command, arg...)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	//cmd.Print(cmd)

	if err := cmd.Run(); err != nil {
		fmt.Fprint(os.Stderr, "Error invoking " + command + ": " + fmt.Sprint(err) + "\n" + stderr.String() + "\n")
		return err, "Error invoking " + command + ": " + err.Error()
	} else {
		fmt.Print(out.String())
	}

	return nil, out.String()
}
