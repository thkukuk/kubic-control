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
	"strings"

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

func ResetNode(nodeName string) (bool, string) {

	ret_success := true
	ret_message := ""

	hostname, err := GetNodeName(nodeName)
	if err != nil {
		return false, err.Error()
	}

	/* ignore if we cannot drain node */
	tools.DrainNode(hostname, "")

	/* Delete the node from the etcd member list if it is on it.
             Else we will can end with a non-functional etcd cluster */
	success, message := tools.ExecuteCmd("etcdctl",
		"--endpoints", "https://localhost:2379",
		"--ca-file",  "/etc/kubernetes/pki/etcd/ca.crt",
		"--cert-file", "/etc/kubernetes/pki/etcd/server.crt",
		"--key-file", "/etc/kubernetes/pki/etcd/server.key",
		"member", "list")
	if success == true {
		var etcd_member_id string

		list := strings.Split (message, "\n")
                for _, entry := range list {
                        if strings.Contains(entry, "name=" + hostname) {
				list := strings.Split (entry, ":");
				etcd_member_id = list[0]

				success, message = tools.ExecuteCmd("etcdctl",
					"--endpoints", "https://localhost:2379",
					"--ca-file",  "/etc/kubernetes/pki/etcd/ca.crt",
					"--cert-file", "/etc/kubernetes/pki/etcd/server.crt",
					"--key-file", "/etc/kubernetes/pki/etcd/server.key",
					"member", "remove", etcd_member_id)
				if success != true {
					ret_message = ret_message + nodeName + ": " + message + " (ignored)\n"
				}
                        }
                }
        }


	/* reset the node. Even if this fails, continue cleanup, but
            report back */
	success, message = tools.ExecuteCmd("salt",  nodeName,
		"cmd.run",  "kubeadm reset --force")
	if success != true {
		ret_message = ret_message + nodeName + ": " + message + " (ignored)\n"
	}

	/* Try some system cleanup, ignore if fails */
        tools.ExecuteCmd("salt", nodeName, "cmd.run",
		"sed -i -e 's|^REBOOT_METHOD=kured|REBOOT_METHOD=auto|g' /etc/transactional-update.conf")
        tools.ExecuteCmd("salt", nodeName, "grains.delkey",  "kubicd")
	tools.ExecuteCmd("salt", nodeName, "cmd.run",
		"\"iptables -F && iptables -t nat -F && iptables -t mangle -F && iptables -X\"")
        tools.ExecuteCmd("salt", nodeName, "cmd.run",  "\"rm -rf /var/lib/etcd/*\"")
        tools.ExecuteCmd("salt", nodeName, "cmd.run",  "\"rm -rf /var/lib/cni/*\"")
        tools.ExecuteCmd("salt", nodeName, "cmd.run",  "\"ip link delete cni0;  ip link delete flannel.1; ip link delete cilium_vxlan\"")
        tools.ExecuteCmd("salt", nodeName, "service.disable",  "kubelet")
	tools.ExecuteCmd("salt", nodeName, "service.stop",  "kubelet")
        tools.ExecuteCmd("salt", nodeName, "service.disable",  "crio")
        tools.ExecuteCmd("salt", nodeName, "service.stop",  "crio")

	/* ignore if we cannot delete the node*/
	success, message = tools.ExecuteCmd("kubectl", "--kubeconfig=/etc/kubernetes/admin.conf",
		"delete",  "node",  hostname)
	if success != true {
		ret_message = ret_message + nodeName + ": " + message + " (ignored)"
	}

	return ret_success, ret_message
}
