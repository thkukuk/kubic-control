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
	"sync"

	pb "github.com/thkukuk/kubic-control/api"
	log "github.com/sirupsen/logrus"
	"github.com/thkukuk/kubic-control/pkg/tools"
)

func RemoveNode(in *pb.RemoveNodeRequest, stream pb.Kubeadm_RemoveNodeServer) error {
	var nodelist []string

	// If we have a list of Nodes, try to find the right node names which have a kubic-worker-node grain.
	if strings.Index(in.NodeNames, ",") >= 0 || strings.Index(in.NodeNames, "[") >= 0 || strings.Compare(in.NodeNames, "*") == 0 {
		var success bool
		var message string

		if strings.Index(in.NodeNames, ",") >= 0 && strings.Index(in.NodeNames, "[")  == -1 {
			success, message = tools.ExecuteCmd("salt", "--out=txt", "-L", in.NodeNames, "grains.get", "kubicd")
		} else {
			success, message = tools.ExecuteCmd("salt", "--out=txt", in.NodeNames, "grains.get",  "kubicd")
		}
		if success != true {
			if err := stream.Send(&pb.StatusReply{Success: false, Message: message}); err != nil {
				return err
			}
			return nil
		}

		list := strings.Split (message, "\n")
		for _, entry := range list {
			if strings.Contains(entry, "'kubic-worker-node'") || strings.Contains(entry, "kubic-master-node") {
				list := strings.Split (entry, ":");
				nodelist = append (nodelist, list[0])
			}
		}
	} else {
		// only one node name to remove
		nodelist = append(nodelist,in.NodeNames)
	}

	nodelistLength := len(nodelist)

	if nodelistLength == 0 {
		if err := stream.Send(&pb.StatusReply{Success: true, Message: "No Nodes found"}); err != nil {
			return err
		}
		return nil
	}

        var wg sync.WaitGroup
        wg.Add(nodelistLength)

        failed := 0
        for i := 0; i < nodelistLength; i++ {
                go func(i int) {
                        defer wg.Done()

                        stream.Send(&pb.StatusReply{Success: true, Message: nodelist[i] + ": removing node..."})
			success, message := ResetNode(nodelist[i])
			if len(message) > 0 {
				if err := stream.Send(&pb.StatusReply{Success: false,
					Message: nodelist[i] + ": " + message}); err != nil {
						log.Errorf("Send message failed: %s", err)
					}
			}
			if success != true {
				failed++
				if err := stream.Send(&pb.StatusReply{Success: false,
					Message: nodelist[i] + ": removal not fully successful, please check logs"}); err != nil {
						log.Errorf("Send message failed: %s", err)
					}
			} else {
				if err := stream.Send(&pb.StatusReply{Success: false,
					Message: nodelist[i] + ": successfully removed"}); err != nil {
						log.Errorf("Send message failed: %s", err)
					}
			}
		}(i)
        }

        wg.Wait()
        if (failed > 0) {
                if err := stream.Send(&pb.StatusReply{Success: false,
			Message: "An error occured during removal of Nodes"}); err != nil {
				return err
			}
        }
        return nil
}
