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

	pb "github.com/thkukuk/kubic-control/api"
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
		// XXX use salt to query node
		useEfi = false
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
		// XXX use salt to query node
		useBareMetal = false
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
		// XXX use salt to query node or report back error to specify on the command line
		entry = "{% set disk = '" + "DEVICE MISSING" + "' %}\n"
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

	// set_perm (OutputDir + "haproxy.cfg")

	if err := stream.Send(&pb.StatusReply{Success: true,
		Message: "Configuration created. Please check \"" + pillarFile + "\" and run install phase."}); err != nil {
			return err
		}
	return nil
}
