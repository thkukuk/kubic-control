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
	"os"
	"strings"
	"encoding/json"

	log "github.com/sirupsen/logrus"
	pb "github.com/thkukuk/kubic-control/api"
	"github.com/thkukuk/kubic-control/pkg/tools"
)

func PrepareConfig(in *pb.PrepareConfigRequest, stream pb.Yomi_PrepareConfigServer) error {

	if err := stream.Send(&pb.StatusReply{Success: true,
		Message: "Prepare salt configuration for Node " + in.Saltnode + " as " + in.Type}); err != nil {
			return err
		}

	if in.Type != "haproxy" {
		if err := stream.Send(&pb.StatusReply{Success: false,
			Message: "Invalid type '" + in.Type + "', valid types are: \"haproxy\""}); err != nil {
				return err
			}
		return nil
	}

	// Get the hardware information from the new node

	if err := stream.Send(&pb.StatusReply{Success: true,
		Message: "Gather hardware informations for Node " + in.Saltnode }); err != nil {
			return err
		}

	// make sure latest modules are used on minion
        success, message := tools.ExecuteCmd("salt", in.Saltnode, "saltutil.sync_all")
        if success != true {
                if err := stream.Send(&pb.StatusReply{Success: false,
                        Message: message}); err != nil {
                                return err
                        }
                return nil
        }

	pillarName := Salt2PillarName(in.Saltnode)
	pillarFile := "/srv/pillar/kubicd/" + pillarName + ".sls"

	f, err := os.Create(pillarFile)
	if err != nil {
		if err2 := stream.Send(&pb.StatusReply{Success: false,
			Message: "Could not create \"" + pillarFile + "\": " + err.Error()}); err2 != nil {
				return err2
			}
		return nil
	}

	_, err = f.WriteString( "# Meta pillar for Yomi\n" +
		"#\n" +
		"# There are some parameters that can be configured and adapted to\n" +
		"# launch a basic Yomi installation:\n" +
		"#\n" +
		"#   * efi = {True, False}\n" +
		"#   * baremetal = {True, False}\n" +
		"#   * disk = {/dev/...}\n" +
		"#   * repo-main = {https://download....}\n" +
		"#\n" +
		"\n")
	if err != nil {
		if err2 := stream.Send(&pb.StatusReply{Success: false,
			Message: "Writing to \"" + pillarFile + "\" failed: " + err.Error()}); err2 != nil {
				return err2
			}
		return nil
	}

	useEfi := false
	if in.Efi == 0 {
		// UEFI or BIOS?
		success, message = tools.ExecuteCmd("salt", "--out=txt", in.Saltnode, "cmd.run",
			"test -f /sys/firmware/efi/systab && echo true || echo false")
		if success != true {
			if err := stream.Send(&pb.StatusReply{Success: false,
				Message: message}); err != nil {
					return err
				}
			return nil
		}
		uefi := strings.Replace(message, "\n","",-1)
		i := strings.Index(uefi,":")+1
		uefi = strings.TrimSpace(uefi[i:])
		log.Info ("UEFI: " + uefi)
		if strings.EqualFold(uefi, "true") {
			useEfi = true
		}
	} else if in.Efi == -1 {
		useEfi = false
	} else {
		useEfi = true
	}
	if useEfi {
		_, err = f.WriteString("{% set efi = True %}\n")
	} else {
		_, err = f.WriteString("{% set efi = False %}\n")
	}
	if err != nil {
		if err2 := stream.Send(&pb.StatusReply{Success: false,
			Message: "Writing to \"" + pillarFile + "\" failed: " + err.Error()}); err2 != nil {
				return err2
			}
		return nil
	}

	useBareMetal := false
	if in.Baremetal == 0 {
		// bare metal or virtualisation?
		success, message = tools.ExecuteCmd("salt", "--out=txt", in.Saltnode, "cmd.run", "systemd-detect-virt")
		if success != true {
			if err := stream.Send(&pb.StatusReply{Success: false,
				Message: message}); err != nil {
					return err
				}
			return nil
		}
		virtualisation := strings.Replace(message, "\n","",-1)
		i := strings.Index(virtualisation,":")+1
		virtualisation = strings.TrimSpace(virtualisation[i:])
		log.Info ("Virtualisation: " + virtualisation)
		if strings.EqualFold(virtualisation, "none") {
			useBareMetal = true
		}
	} else if in.Baremetal == -1 {
		useBareMetal = false
	} else {
		useBareMetal = true
	}
	if useBareMetal {
		_, err = f.WriteString("{% set baremetal = True %}\n")
	} else {
		_, err = f.WriteString("{% set baremetal = False %}\n")
	}
	if err != nil {
		if err2 := stream.Send(&pb.StatusReply{Success: false,
			Message: "Writing to \"" + pillarFile + "\" failed: " + err.Error()}); err2 != nil {
				return err2
			}
		return nil
	}

	entry := ""
	if len(in.Disk) > 0 {
		entry = "{% set disk = '" + in.Disk + "' %}\n"
	} else {
		success, message = tools.ExecuteCmd("salt", "--out=json", in.Saltnode, "devices.hwinfo", "disk")
		if success != true {
			if err := stream.Send(&pb.StatusReply{Success: false,
				Message: message}); err != nil {
					return err
				}
			return nil
		}
		var hwinfo_all map[string]interface{}
		err = json.Unmarshal([]byte(message), &hwinfo_all)
		if err != nil {
			if err2 := stream.Send(&pb.StatusReply{Success: false,
				Message: "Detecting disks failed: " + err.Error()}); err2 != nil {
					return err2
				}
			return nil
		}
		hwinfo_node := hwinfo_all[in.Saltnode].(map[string]interface{})
		hwinfo := hwinfo_node["hwinfo"].(map[string]interface{})
		hwinfo_disk := hwinfo["disk"].(map[string]interface{})
		if len (hwinfo_disk) != 1 {
			message = "Found more than one disk:\n"
			for key, value := range hwinfo_disk {
				message = message + "- " + key + " (" + value.(string) +")\n"
			}
			if err := stream.Send(&pb.StatusReply{Success: false,
				Message: message}); err != nil {
					return err
				}
			return nil
		}

		// XXX we have exactly one key, no easier way?
		for key, _ := range hwinfo_disk {
			entry = "{% set disk = '" + key + "' %}\n"
		}
	}

	if len(in.Repo) > 0 {
		entry = entry + "{% set repo_main = '" + in.Repo + "' %}\n"
	} else {
		entry = entry + "{% set repo_main = 'http://download.opensuse.org/tumbleweed/repo/oss' %}"
	}

	_, err = f.WriteString(entry +
		"\n" +
		"{% include \"kubicd/_haproxy.sls\" %}\n\n")
	if err != nil {
		if err2 := stream.Send(&pb.StatusReply{Success: false,
			Message: "Writing to \"" + pillarFile + "\" failed: " + err.Error()}); err2 != nil {
				return err2
			}
		return nil
	}

	if err := f.Close(); err != nil {
		if err2 := stream.Send(&pb.StatusReply{Success: false,
			Message: "Closing \"" + pillarFile + "\" failed: " + err.Error()}); err2 != nil {
				return err2
			}
		return nil
	}

	// XXX set_perm (OutputDir + "haproxy.cfg")

	if err := stream.Send(&pb.StatusReply{Success: true,
		Message: "Configuration created. Please check \"" + pillarFile + "\" and run install phase."}); err != nil {
			return err
		}
	return nil
}
