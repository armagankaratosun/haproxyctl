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
	"strconv"

	"github.com/spf13/cobra"
)

// DeleteServersCmd represents "delete server"
var DeleteServersCmd = &cobra.Command{
	Use:   "server <backend_name> <server_name>",
	Short: "Delete a specific HAProxy server from a backend",
	Long: `This command deletes a server from a specific backend.

Example:
  haproxyctl delete server mybackend myserver`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		backendName := args[0]
		serverName := args[1]
		deleteServer(backendName, serverName)
	},
}

// deleteServer handles deletion of a server from a backend
func deleteServer(backendName, serverName string) {
	version, err := utils.GetConfigurationVersion()
	if err != nil {
		log.Fatalf("Failed to fetch HAProxy configuration version: %v", err)
	}

	endpoint := fmt.Sprintf("/services/haproxy/configuration/backends/%s/servers/%s", backendName, serverName)
	_, err = utils.SendRequest("DELETE", endpoint, map[string]string{"version": strconv.Itoa(version)}, nil)
	if err != nil {
		log.Fatalf("Failed to delete server '%s' in backend '%s': %v", serverName, backendName, err)
	}

	fmt.Printf("Server '%s' in backend '%s' deleted successfully.\n", serverName, backendName)
}
