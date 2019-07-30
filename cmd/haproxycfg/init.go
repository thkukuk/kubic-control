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

package main

import (
	"fmt"
	"os"

        "github.com/spf13/cobra"
)

var (
	output_dir = "/etc/haproxy"
)

func InitializeConfigCmd() *cobra.Command {
        var subCmd = &cobra.Command {
                Use:   "initialize <loadbalancer DNS name> <first master IP>",
                Short: "Create initial haproxy.cfg overwriting existing one",
                Run: initializeConfig,
                Args: cobra.ExactArgs(2),
        }

	subCmd.PersistentFlags().StringVar(&output_dir, "dir", output_dir, "Directory, in which haproxy.cfg should be written")

        return subCmd
}

func initializeConfig (cmd *cobra.Command, args []string) {

	lb_dns_name := args[0]
	apiserver1 := args[1]

	if len(output_dir) > 0 && output_dir[len(output_dir)-1:] != "/" {
		output_dir = output_dir + "/"
	}

	f, err := os.Create(output_dir + "haproxy.cfg")
	if err != nil {
                fmt.Fprintf(os.Stderr, "Could not create \"" + output_dir + "haproxy.cfg\": %v", err)
		os.Exit(1)
        }

	_, err = f.WriteString("global\n" +
		"  log /dev/log daemon\n" +
		"  maxconn 32768\n" +
		"  chroot /var/lib/haproxy\n" +
		"  user haproxy\n" +
		"  group haproxy\n" +
		"  daemon\n" +
		"  stats socket /var/lib/haproxy/stats user haproxy group haproxy mode 0640 level operator\n" +
		"  tune.bufsize 32768\n" +
		"  tune.ssl.default-dh-param 2048\n" +
		"  ssl-default-bind-ciphers ALL:!aNULL:!eNULL:!EXPORT:!DES:!3DES:!MD5:!PSK:!RC4:!ADH:!LOW@STRENGTH\n" +
		"\n" +
		"defaults\n" +
		"  log     global\n" +
		"  mode    http\n" +
		"  option  log-health-checks\n" +
		"  option  log-separate-errors\n" +
		"  option  dontlog-normal\n" +
		"  option  dontlognull\n" +
		"  option  httplog\n" +
		"  option  socket-stats\n" +
		"  retries 3\n" +
		"  option  redispatch\n" +
		"  maxconn 10000\n" +
		"  timeout connect     5s\n" +
		"  timeout client     50s\n" +
		"  timeout server    450s\n" +
		"\n" +
		"listen stats\n" +
		"  bind localhost:80\n" +
		"  stats enable\n" +
		"  stats uri     /stats\n" +
		"  stats refresh 5s\n" +
		"\n")
	if err != nil {
                fmt.Fprintf(os.Stderr, "Writing to \"" + output_dir + "haproxy.cfg\" failed: %v", err)
		os.Exit(1)
        }

	// XXX make writing frontend and backend an own function to add this sections later to an existing file
	_, err = f.WriteString("frontend k8s-api\n" +
		"    bind " + lb_dns_name + ":6443\n" +
		"    bind localhost:6443\n" +
		"    mode tcp\n" +
		"    option tcplog\n" +
		"    default_backend k8s-api\n" +
		"\n" +
		"backend k8s-api\n" +
		"    mode tcp\n" +
		"    option tcp-check\n" +
		"    balance roundrobin\n" +
		"    default-server inter 10s downinter 5s rise 2 fall 2 slowstart 60s maxconn 250 maxqueue 256 weight 100\n" +
		"    server apiserver1 " + apiserver1 + ":6443 check\n\n")
	if err != nil {
                fmt.Fprintf(os.Stderr, "Writing to \"" + output_dir + "haproxy.cfg\" failed: %v", err)
		os.Exit(1)
        }

	f.Close()

	fmt.Printf("haproxy.cfg created\n")
}
