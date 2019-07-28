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
        "os"
	"path/filepath"

        "github.com/thkukuk/kubic-control/pkg/tools"
)

func removeContents(dir string) error {
    files, err := filepath.Glob(filepath.Join(dir, "*"))
    if err != nil {
        return err
    }
    for _, file := range files {
        err = os.RemoveAll(file)
        if err != nil {
            return err
        }
    }
    return nil
}

func ResetMaster() (bool, string) {

	success, message :=  tools.ExecuteCmd("kubeadm", "reset", "--force")

	// cleanup behind kubeadm
	removeContents("/var/lib/etcd")
	removeContents("/var/lib/cni")

	os.Remove("/var/lib/kubic-control/control-plane.conf")
	os.Remove("/var/lib/kubic-control/k8s-yaml.conf")

	tools.ExecuteCmd("systemctl", "disable", "--now", "crio")
	tools.ExecuteCmd("systemctl", "disable", "--now", "kubelet")

	return success, message
}
