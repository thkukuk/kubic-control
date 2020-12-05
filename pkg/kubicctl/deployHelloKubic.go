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

package kubicctl

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	pb "github.com/thkukuk/kubic-control/api"
)

var (
	service_type = "NodePort"
	arg_lbip     = ""
)

func DeployHelloKubicCmd() *cobra.Command {
	var subCmd = &cobra.Command{
		Use:   "hello-kubic",
		Short: "Deploy hello-kubic",
		Run:   deployHelloKubic,
		Args:  cobra.ExactArgs(0),
	}

	subCmd.PersistentFlags().StringVarP(&service_type, "type", "t", service_type, "Type for this service: NodePort or LoadBalancer")
	subCmd.PersistentFlags().StringVarP(&arg_lbip, "ip", "i", arg_lbip, "LoadBalancer IP")

	return subCmd
}

func deployHelloKubic(cmd *cobra.Command, args []string) {

	// Set up a connection to the server.
	conn, err := CreateConnection()
	if err != nil {
		return
	}
	defer conn.Close()

	c := pb.NewDeployClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var arg string

	if strings.EqualFold(service_type, "NodePort") {
		arg = "NodePort"
	} else if strings.EqualFold(service_type, "LoadBalancer") {
		// If we use loadbalancer with a prefered IP, only
		// transfer the IP, we have no second argument.
		if len(arg_lbip) > 0 {
			arg = arg_lbip
		} else {
			arg = "LoadBalancer"
		}
	}

	r, err := c.DeployKustomize(ctx,
		&pb.DeployKustomizeRequest{Service: "hello-kubic", Argument: arg})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize: %v\n", err)
		os.Exit(1)
	}
	if r.Success {
		if len(output) > 0 && output != "stdout" {
			message := []byte(r.Message)
			err := ioutil.WriteFile(output, message, 0600)
			if err != nil {
				fmt.Fprintf(os.Stderr,
					"Error writing '%s': %v\n", output, err)
				os.Exit(1)
			}
		} else {
			fmt.Printf(r.Message)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Couldn't deploy hello-kubic: %s\n",
			r.Message)
		os.Exit(1)
	}
}
