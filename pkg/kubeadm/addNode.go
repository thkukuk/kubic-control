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
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/thkukuk/kubic-control/pkg/tools"
)

var (
	joincmd = ""
	token_create_time time.Time
)

func AddNode(nodeNames string) (bool, string) {
	// XXX Check if node isn't already part of the kubernetes cluster

	// If the join command is older than 23 hours, generate a new one. Else re-use the old one.
	if time.Since(token_create_time).Hours() > 23 {
		log.Info("Token to join nodes too old, creating new one")
		success, token := tools.ExecuteCmd("kubeadm", "token", "create", "--print-join-command")
		if success != true {
			return success, joincmd
		}
		joincmd = strings.TrimSuffix(token, "\n")
		token_create_time = time.Now()
	}

	// Differentiate between 'name1,name2' and 'name[1,2]'
	if strings.Index(nodeNames, ",") >= 0 && strings.Index(nodeNames, "[") == -1 {
		success, message := tools.ExecuteCmd("salt", "-L", nodeNames, "service.start", "crio")
		if success != true {
			return success, message
		}
		success, message = tools.ExecuteCmd("salt", "-L", nodeNames, "service.enable", "crio")
		if success != true {
			return success, message
		}
		success, message = tools.ExecuteCmd("salt", "-L", nodeNames, "service.start", "kubelet")
		if success != true {
			return success, message
		}
		success, message = tools.ExecuteCmd("salt", "-L", nodeNames, "service.enable", "kubelet")
		if success != true {
			return success, message
		}
		success, message = tools.ExecuteCmd("salt", "-L", nodeNames, "cmd.run", "\"" + joincmd + "\"")
		if success != true {
			return success, message
		}
		success, message = tools.ExecuteCmd("salt", "-L", nodeNames, "grains.append", "kubicd", "kubic-worker-node")
		if success != true {
			return success, message
		}
		success, message = tools.ExecuteCmd("salt", "-L", nodeNames, "cmd.run", "if [ -f /etc/transactional-update.conf ]; then grep -q ^REBOOT_METHOD= /etc/transactional-update.conf && sed -i -e 's|REBOOT_METHOD=.*|REBOOT_METHOD=kured|g' /etc/transactional-update.conf || echo REBOOT_METHOD=kured >> /etc/transactional-update.conf ; else echo REBOOT_METHOD=kured > /etc/transactional-update.conf ; fi")
		if success != true {
			return success, message
		}
	} else {
		success, message := tools.ExecuteCmd("salt", nodeNames, "service.start", "crio")
		if success != true {
			return success, message
		}
		success, message = tools.ExecuteCmd("salt", nodeNames, "service.enable", "crio")
		if success != true {
			return success, message
		}
		success, message = tools.ExecuteCmd("salt", nodeNames, "service.start", "kubelet")
		if success != true {
			return success, message
		}
		success, message = tools.ExecuteCmd("salt", nodeNames, "service.enable", "kubelet")
		if success != true {
			return success, message
		}
		success, message = tools.ExecuteCmd("salt",  nodeNames, "cmd.run",  "\"" + joincmd + "\"")
		if success != true {
			return success, message
		}
		success, message = tools.ExecuteCmd("salt", nodeNames, "grains.append", "kubicd", "kubic-worker-node")
		if success != true {
			return success, message
		}
		// Configure transactinal-update
		success, message = tools.ExecuteCmd("salt", nodeNames, "cmd.run", "if [ -f /etc/transactional-update.conf ]; then grep -q ^REBOOT_METHOD= /etc/transactional-update.conf && sed -i -e 's|REBOOT_METHOD=.*|REBOOT_METHOD=kured|g' /etc/transactional-update.conf || echo REBOOT_METHOD=kured >> /etc/transactional-update.conf ; else echo REBOOT_METHOD=kured > /etc/transactional-update.conf ; fi")
		if success != true {
			return success, message
		}
	}

	return true, ""
}
