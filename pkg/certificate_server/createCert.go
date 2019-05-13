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

package certificate

import (
        "os/exec"
        "fmt"
        "bytes"

	log "github.com/sirupsen/logrus"
	pb "github.com/thkukuk/kubic-control/api"
)

var PKI_dir = "/etc/kubicd/pki"

func ExecuteCmd(command string, arg ...string) (bool,string) {
        var out bytes.Buffer
        var stderr bytes.Buffer

        cmd := exec.Command(command, arg...)
        cmd.Stdout = &out
        cmd.Stderr = &stderr

        //log.Info(cmd)

        if err := cmd.Run(); err != nil {
                log.Error("Error invoking " + command + ": " + fmt.Sprint(err) + "\n" + stderr.String())
                return false, "Error invoking " + command + ": " + err.Error()
        } else {
                log.Info(out.String())
        }

        return true, out.String()
}


func CreateUser (pki_dir string, cn string) (bool, string) {
	return ExecuteCmd("certstrap", "--depot-path", pki_dir,
		"request-cert", "--common-name", cn, "--passphrase", "")
}

func SignUser (pki_dir string, cn string) (bool, string) {
        return ExecuteCmd("certstrap", "--depot-path", pki_dir, "sign",
		cn, "--CA", "Kubic-Control-CA")
}

func CreateCert(in *pb.CreateCertRequest) (bool, string, string, string) {

	user := in.Name

        success, message := CreateUser(PKI_dir, user)
        if success != true {
                return success, message, "", ""
        }
        success, message = SignUser(PKI_dir, user)
        if  success != true {
                return success, message, "", ""
        }

	return true, "Not fully implemented", "", ""
}
