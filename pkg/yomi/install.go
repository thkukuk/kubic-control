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
	pb "github.com/thkukuk/kubic-control/api"
	"github.com/thkukuk/kubic-control/pkg/tools"
)

func Install(in *pb.InstallRequest, stream pb.Yomi_InstallServer) error {

	if err := stream.Send(&pb.StatusReply{Success: true,
		Message: "Starting installation of " + in.Saltnode}); err != nil {
		return err
	}

	pillarName := Salt2PillarName(in.Saltnode)
	pillarFile := "/srv/pillar/kubicd/" + pillarName + ".sls"

	exists, _ := tools.Exists(pillarFile)
	if !exists {
		if err := stream.Send(&pb.StatusReply{Success: false,
			Message: "No pillar data found, prepare config step not run?"}); err != nil {
			return err
		}
		return nil
	}

	// make sure latest modules are used on minion
	success, message := tools.ExecuteCmd("salt", "--module-executors='direct_call'", in.Saltnode, "saltutil.sync_all")
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: false,
			Message: message}); err != nil {
			return err
		}
		return nil
	}

	// wipe harddisk, else salt will not re-create them
	success, message = tools.ExecuteCmd("salt", "--module-executors='direct_call'", in.Saltnode, "state.apply", "yomi.storage.wipe")
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: false,
			Message: message}); err != nil {
			return err
		}
		return nil
	}

	// Do final installation
	success, message = tools.ExecuteCmd("salt", "--module-executors='direct_call'", in.Saltnode, "state.sls", "yomi.installer")
	if success != true {
		if err := stream.Send(&pb.StatusReply{Success: false,
			Message: message}); err != nil {
			return err
		}
		return nil
	}

	if err := stream.Send(&pb.StatusReply{Success: true,
		Message: "Node successful installed!"}); err != nil {
		return err
	}
	return nil
}
