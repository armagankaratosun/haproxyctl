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
	"haproxyctl/utils"
	"log"

	"github.com/spf13/cobra"
)

// DescribeServersCmd represents "describe server"
var DescribeServersCmd = &cobra.Command{
	Use:   "server <backend_name> <server_name>",
	Short: "Describe a specific HAProxy server in a backend",
	Long: `Retrieve detailed information about a server inside a specific backend.

Example:
  haproxyctl describe server mybackend myserver`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		backendName := args[0]
		serverName := args[1]
		describeServer(backendName, serverName)
	},
}

// describeServer fetches and prints details of a server within a backend
func describeServer(backendName, serverName string) {
	endpoint := fmt.Sprintf("/services/haproxy/configuration/backends/%s/servers/%s", backendName, serverName)

	server, err := utils.GetResource(endpoint)
	if err != nil {
		log.Fatalf("Failed to fetch server '%s' in backend '%s': %v", serverName, backendName, err)
	}

	utils.PrintResourceDescription("Server", server, serverDescriptionSections(), nil)
}

// serverDescriptionSections defines sections for server description output
func serverDescriptionSections() map[string][]string {
	return map[string][]string{
		"basic":    {"name", "address", "port", "weight"},
		"health":   {"check", "check_alpn", "check_ssl", "inter", "rise", "fall"},
		"advanced": {"maxconn", "ssl", "verify", "sni"},
	}
}
