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
	"gopkg.in/ini.v1"
        pb "github.com/thkukuk/kubic-control/api"
	log "github.com/sirupsen/logrus"
	"github.com/thkukuk/kubic-control/pkg/tools"
)

func GetStatus(in *pb.Empty, stream pb.Kubeadm_GetStatusServer, kubicdVersion string) error {

	if err := stream.Send(&pb.StatusReply{Success: true,
		Message: "Kubicd version: " + kubicdVersion}); err != nil {
			log.Errorf("Send message failed: %s", err)
			return err
		}
	_, message := tools.GetKubeadmVersion("") // XXX needs better handling, per master via salt.
	if err := stream.Send(&pb.StatusReply{Success: true,
		Message: "kubeadm version: " + message}); err != nil {
			log.Errorf("Send message failed: %s", err)
			return err
		}

	// Standard yaml files
        cfg, err := ini.Load("/var/lib/kubic-control/k8s-yaml.conf")
        if err != nil {
		if err := stream.Send(&pb.StatusReply{Success: false,
			Message: "Cannot load k8s-yaml.conf: " + err.Error()}); err != nil {
				log.Errorf("Send message failed: %s", err)
				return err
		}
        } else {

		keys := cfg.Section("").KeyStrings()

		if len(keys) > 0 {
			if err := stream.Send(&pb.StatusReply{Success: true,
				Message: "Status of deployed daemonsets (yaml):"}); err != nil {
					log.Errorf("Send message failed: %s", err)
					return err
				}
		}
		for _, key := range keys {
			value := cfg.Section("").Key(key).String()
			hash, _ := tools.Sha256sum_f(key)
                        if hash != value {
				if err := stream.Send(&pb.StatusReply{Success: true,
					Message: "- " + key + ": newer version available"}); err != nil {
						log.Errorf("Send message failed: %s", err)
						return err
					}
                        } else {
				if err := stream.Send(&pb.StatusReply{Success: true,
					Message: "- " + key + ": up to date"}); err != nil {
						log.Errorf("Send message failed: %s", err)
						return err
					}

                        }
                }
        }

	// kustomize
        cfg, err = ini.Load("/var/lib/kubic-control/k8s-kustomize.conf")
        if err != nil {
		if err := stream.Send(&pb.StatusReply{Success: false,
			Message: "Cannot load k8s-yaml.conf: " + err.Error()}); err != nil {
				log.Errorf("Send message failed: %s", err)
				return err
		}
        } else {

		keys := cfg.Section("").KeyStrings()

		if len(keys) > 0 {
			if err := stream.Send(&pb.StatusReply{Success: true,
				Message: "Status of deployed daemonsets (kustomize):"}); err != nil {
					log.Errorf("Send message failed: %s", err)
					return err
				}
		}
		for _, key := range keys {
			value := cfg.Section("").Key(key).String()
			_, output := tools.ExecuteCmd("kustomize", "build",
				"/var/lib/kubic-control/kustomize/" + key + "/overlay")
			hash, _ := tools.Sha256sum_b(output)
                        if hash != value {
				if err := stream.Send(&pb.StatusReply{Success: true,
					Message: "- " + key + ": newer version available"}); err != nil {
						log.Errorf("Send message failed: %s", err)
						return err
					}
                        } else {
				if err := stream.Send(&pb.StatusReply{Success: true,
					Message: "- " + key + ": up to date"}); err != nil {
						log.Errorf("Send message failed: %s", err)
						return err
					}

                        }
                }
        }

	return nil
}
