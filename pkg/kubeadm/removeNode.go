// Copyright 2019, 2020 Thorsten Kukuk
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

	log "github.com/sirupsen/logrus"
	pb "github.com/thkukuk/kubic-control/api"
	"github.com/thkukuk/kubic-control/pkg/tools"
)

// output_stream types takes bool and string, returns nothing.
type OutputStream func(bool, string)

var output_stream pb.Kubeadm_RemoveNodeServer

func RemoveNodeOutput(success bool, message string) {
	if err := output_stream.Send(&pb.StatusReply{Success: success,
		Message: message}); err != nil {
		log.Errorf("Send message failed: %s", err)
	}
}

func RemoveNode(in *pb.RemoveNodeRequest, stream pb.Kubeadm_RemoveNodeServer) error {
	var nodelist []string
	output_stream = stream

	// If we have a list of Nodes, try to find the right node names which
	// have a kubic-worker-node or kubic-master-node grain.
	if strings.Index(in.NodeNames, ",") >= 0 || strings.Index(in.NodeNames, "[") >= 0 || strings.Compare(in.NodeNames, "*") == 0 {
		var success bool
		var message string

		if strings.Index(in.NodeNames, ",") >= 0 && strings.Index(in.NodeNames, "[") == -1 {
			success, message = tools.ExecuteCmd("salt", "--out=txt", "-L", in.NodeNames, "grains.get", "kubicd")
		} else {
			success, message = tools.ExecuteCmd("salt", "--out=txt", in.NodeNames, "grains.get", "kubicd")
		}
		if success != true {
			if err := stream.Send(&pb.StatusReply{Success: false, Message: message}); err != nil {
				return err
			}
			return nil
		}

		list := strings.Split(message, "\n")
		for _, entry := range list {
			if strings.Contains(entry, "'kubic-worker-node'") || strings.Contains(entry, "kubic-master-node") {
				list := strings.Split(entry, ":")
				nodelist = append(nodelist, list[0])
			}
		}
	} else {
		// only one node name to remove
		nodelist = append(nodelist, in.NodeNames)
	}

	nodelistLength := len(nodelist)

	if nodelistLength == 0 {
		if err := stream.Send(&pb.StatusReply{Success: true, Message: "No Nodes found"}); err != nil {
			return err
		}
		return nil
	}

	haproxy_salt := Read_Cfg("control-plane.conf", "loadbalancer_salt")
	var wg sync.WaitGroup
	wg.Add(nodelistLength)

	failed := 0
	for i := 0; i < nodelistLength; i++ {
		go func(i int) {
			defer wg.Done()

			stream.Send(&pb.StatusReply{Success: true, Message: nodelist[i] + ": start node removal..."})

			// If loadbalancer is known, remove from haproxy
			if len(haproxy_salt) > 0 {
				stream.Send(&pb.StatusReply{Success: true, Message: nodelist[i] + ": removing node from haproxy loadbalancer..."})
				success, message := tools.ExecuteCmd("salt", haproxy_salt, "cmd.run", "haproxycfg server remove "+nodelist[i])
				if success != true {
					if err := stream.Send(&pb.StatusReply{Success: false, Message: nodelist[i] + ": " + message}); err != nil {
						log.Errorf("Send message failed: %s", err)
					}
					failed++ // XXX try to detect type: ignore for worker
				}
			}

			success, message := ResetNode(nodelist[i], RemoveNodeOutput)
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
	if failed > 0 {
		if err := stream.Send(&pb.StatusReply{Success: false,
			Message: "An error occured during removal of Nodes"}); err != nil {
			return err
		}
	}
	return nil
}
