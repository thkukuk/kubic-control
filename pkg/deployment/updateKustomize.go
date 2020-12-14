// Copyright 2020 Thorsten Kukuk
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

package deployment

import (
	"os"

	"github.com/thkukuk/kubic-control/pkg/tools"
	"gopkg.in/ini.v1"
)

func UpdateKustomize(service string) (bool, string) {

	retval, message := tools.ExecuteCmd("kustomize", "build",
		StateDir+"/kustomize/"+service+"/overlay")
	if retval != true {
		return false, message
	}

	f, err := os.Create(StateDir + "/kustomize/" + service + "/" + service + ".yaml")
	if err != nil {
		return false, err.Error()
	}
	defer f.Close()
	_, err = f.WriteString(message)
	if err != nil {
		return false, err.Error()
	}
	f.Close()

	retval, message = tools.ExecuteCmd("kubectl",
		"--kubeconfig=/etc/kubernetes/admin.conf", "apply", "-f",
		StateDir+"/kustomize/"+service+"/"+service+".yaml")
	if retval != true {
		return false, message
	}

	cfg, err := ini.LooseLoad(StateDir + "/k8s-kustomize.conf")
	if err != nil {
		return false, "Cannot load k8s-kustomize.conf: " + err.Error()
	}

	result, err := tools.Sha256sum_f(StateDir + "/kustomize/" + service + "/" + service + ".yaml")
	cfg.Section("").Key(service).SetValue(result)
	err = cfg.SaveTo(StateDir + "/k8s-kustomize.conf")
	if err != nil {
		return false, "Cannot write k8s-kustomize.conf: " + err.Error()
	}

	return true, ""
}
