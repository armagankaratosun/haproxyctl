/*
Copyright © 2025 Armagan Karatosun

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package servers

import (
	"fmt"
	"haproxyctl/internal"
	"log"

	"github.com/spf13/cobra"
)

// GetServersCmd represents "get servers"
var GetServersCmd = &cobra.Command{
	Use:     "servers <backend_name> [server_name]",
	Aliases: []string{"server"},
	Short:   "List servers in a backend, or fetch details for a specific server",
	Long: `Retrieve the list of servers in a backend, or details of a single server.

Examples:
  haproxyctl get servers mybackend
  haproxyctl get servers mybackend myserver`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		backendName := args[0]
		var serverName string
		if len(args) > 1 {
			serverName = args[1]
		}
		getServers(cmd, backendName, serverName)
	},
}

// getServers fetches the list of servers or a specific server from a backend
func getServers(cmd *cobra.Command, backendName, serverName string) {
	endpoint := fmt.Sprintf("/services/haproxy/configuration/backends/%s/servers", backendName)
	if serverName != "" {
		endpoint += "/" + serverName
	}

	data, err := internal.SendRequest("GET", endpoint, nil, nil)
	if err != nil {
		log.Fatalf("Failed to fetch server(s) from backend '%s': %v", backendName, err)
	}

	outputFormat := internal.GetFlagString(cmd, "output")
	internal.FormatOutput(string(data), outputFormat)
}

func init() {
	// Inherit the global -o flag for output formatting
	GetServersCmd.Flags().StringP("output", "o", "", "Output format: yaml or json")
}
