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
	pb "github.com/thkukuk/kubic-control/api"
	"github.com/thkukuk/kubic-control/pkg/tools"
)

func DestroyMaster(in *pb.Empty, stream pb.Kubeadm_DestroyMasterServer) error {
	success, message := ResetMaster()
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: true, Message: message + " (ignored)"}); err != nil {
			return err
		}
		// ignore error
	}
	// Try some system cleanup, ignore if fails
	tools.ExecuteCmd("/bin/sh", "-c", "sed -i -e 's|^REBOOT_METHOD=kured|REBOOT_METHOD=auto|g' /etc/transactional-update.conf")
	success, message = tools.ExecuteCmd("/bin/sh", "-c", "iptables -F && iptables -t nat -F && iptables -t mangle -F && iptables -X")
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: true, Message: "Warning: removal of iptables failed."}); err != nil {
			return err
		}
	}
	tools.ExecuteCmd("/bin/sh", "-c", "ip link delete cni0;  ip link delete flannel.1; ip link delete cilium_vxlan")

	return nil
}
