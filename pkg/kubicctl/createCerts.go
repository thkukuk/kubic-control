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

package kubicctl

import (
	"context"
        "time"
        "fmt"
	"os"

        "github.com/spf13/cobra"
	pb "github.com/thkukuk/kubic-control/api"
)

func CreateCertsCmd() *cobra.Command {
        var subCmd = &cobra.Command {
                Use:   "create <user>",
                Short: "Cerate certificate for an user",
                Run: createCerts,
                Args: cobra.ExactArgs(1),
        }

        return subCmd
}

func createCerts (cmd *cobra.Command, args []string) {
	user := args[0]

	// Set up a connection to the server.
        conn, err := CreateConnection()
        if err != nil {
                return
        }
        defer conn.Close()

        c := pb.NewCertificateClient(conn)

        ctx, cancel := context.WithTimeout(context.Background(), time.Second)
        defer cancel()

	r, err := c.CreateCert(ctx, &pb.CreateCertRequest{Name: user})
        if err != nil {
                fmt.Fprintf(os.Stderr, "Could not initialize: %v\n", err)
                os.Exit(1)
        }
        if r.Success {
               // if len(output) > 0 && output != "stdout" {
               //         message:=[]byte(r.Message)
               //         err := ioutil.WriteFile(output, message, 0600)
               //         if err != nil {
               //                 fmt.Fprintf(os.Stderr, "Error writing '%s': %v\n", output, err)
               //                 os.Exit(1)
               //         }
              //  } else {
                        fmt.Printf(r.Message)
              //  }
        } else {
                fmt.Fprintf(os.Stderr, "Couldn't create certificate for %s: %s\n", user, r.Message)
                os.Exit(1)
        }
}
