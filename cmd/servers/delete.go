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

// Package servers provides commands to manage HAProxy backend servers.
package servers

import (
	"fmt"
	"haproxyctl/internal"
	"log"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

// DeleteServersCmd represents "delete server".
var DeleteServersCmd = &cobra.Command{
	Use:     "server <backend_name> <server_name>",
	Aliases: []string{"servers"},
	Short:   "Delete a specific HAProxy server from a backend",
	Long: `This command deletes a server from a specific backend.

Example:
  haproxyctl delete server mybackend myserver`,
	Args: cobra.ExactArgs(serverArgsTwo),
	Run: func(_ *cobra.Command, args []string) {
		backendName := args[0]
		serverName := args[1]
		deleteServer(backendName, serverName)
	},
}

// deleteServer handles deletion of a server from a backend.
func deleteServer(backendName, serverName string) {
	if err := DeleteServer(backendName, serverName); err != nil {
		log.Fatalf("Failed to delete server '%s' in backend '%s': %v", serverName, backendName, err)
	}
}

// DeleteServer removes a server from a backend via the Data Plane API.
// This is shared between the CLI command and higher-level workflows
// such as backend editing.
func DeleteServer(backendName, serverName string) error {
	version, err := internal.GetConfigurationVersion()
	if err != nil {
		return fmt.Errorf("failed to fetch HAProxy configuration version: %w", err)
	}

	endpoint := fmt.Sprintf("/services/haproxy/configuration/backends/%s/servers/%s", backendName, serverName)
	_, err = internal.SendRequest("DELETE", endpoint, map[string]string{"version": strconv.Itoa(version)}, nil)
	if err != nil {
		return fmt.Errorf("failed to delete server '%s' in backend '%s': %w", serverName, backendName, err)
	}

	if _, err := fmt.Fprintf(os.Stdout, "Server '%s' in backend '%s' deleted successfully.\n", serverName, backendName); err != nil {
		log.Printf("warning: failed to write server deleted message: %v", err)
	}
	return nil
}
