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
	"log"
	"strconv"

	"haproxyctl/internal"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// CreateServersCmd represents "create servers".
var CreateServersCmd = &cobra.Command{
	Use:     "servers <backend_name> <server_name>",
	Aliases: []string{"server"},
	Short:   "Create a server within a backend",
	Long: `Create a new server inside a specific backend.

Examples:
  haproxyctl create servers mybackend myserver \
    --address 10.0.0.1 \
    --port 80 \
    --weight 100`,
	Args: cobra.ExactArgs(serverArgsTwo),
	Run: func(cmd *cobra.Command, args []string) {
		backendName := args[0]
		serverName := args[1]

		var server ServerConfig
		server.LoadFromFlags(cmd, backendName, serverName)

		if err := server.Validate(); err != nil {
			log.Fatalf("Invalid server configuration: %v", err)
		}

		outputFormat := internal.GetFlagString(cmd, "output")
		dryRun := internal.GetFlagBool(cmd, "dry-run")

		if err := CreateServer(server, outputFormat, dryRun); err != nil {
			log.Fatalf("Failed to create server: %v", err)
		}
	},
}

// CreateServer is the shared function to create a server via API.
// Used by both `create servers` and `create backends --server=...`.
func CreateServer(server ServerConfig, outputFormat string, dryRun bool) error {
	if outputFormat != "" || dryRun {
		// For structured formats, preview the actual API payload
		// (with ssl encoded as an enum). For the default case
		// (no -o), render a YAML view of the richer ServerConfig.
		if outputFormat == "" {
			outputFormat = internal.OutputFormatYAML
			internal.FormatOutput(server, outputFormat)
		} else {
			internal.FormatOutput(server.toPayload(), outputFormat)
		}
		if dryRun {
			internal.PrintDryRun()
		}
		return nil
	}

	if err := server.NormalizeParent(); err != nil {
		return fmt.Errorf("invalid server configuration: %w", err)
	}

	version, err := internal.GetConfigurationVersion()
	if err != nil {
		return fmt.Errorf("failed to fetch HAProxy configuration version: %w", err)
	}

	endpoint := fmt.Sprintf("/services/haproxy/configuration/backends/%s/servers", server.Parent)

	_, err = internal.SendRequest("POST", endpoint,
		map[string]string{"version": strconv.Itoa(version)},
		server.toPayload(),
	)
	if err != nil {
		return fmt.Errorf("failed to create server '%s': %w", server.Name, err)
	}

	displayName := fmt.Sprintf("%s/%s", server.Parent, server.Name)
	internal.PrintStatus("Server", displayName, internal.ActionCreated)
	return nil
}

// UpdateServer updates an existing server definition in a backend via
// the HAProxy Data Plane API v3.
func UpdateServer(server ServerConfig) error {
	if err := server.NormalizeParent(); err != nil {
		return fmt.Errorf("invalid server configuration: %w", err)
	}

	version, err := internal.GetConfigurationVersion()
	if err != nil {
		return fmt.Errorf("failed to fetch HAProxy configuration version: %w", err)
	}

	endpoint := fmt.Sprintf(
		"/services/haproxy/configuration/backends/%s/servers/%s",
		server.Parent,
		server.Name,
	)

	_, err = internal.SendRequest("PUT", endpoint,
		map[string]string{"version": strconv.Itoa(version)},
		server.toPayload(),
	)
	if err != nil {
		return fmt.Errorf("failed to update server '%s' in backend '%s': %w", server.Name, server.Parent, err)
	}

	displayName := fmt.Sprintf("%s/%s", server.Parent, server.Name)
	internal.PrintStatus("Server", displayName, internal.ActionConfigured)
	return nil
}

// CreateServerFromFile handles creating a server from a YAML file.
func CreateServerFromFile(data []byte) error {
	var server ServerConfig
	if err := yaml.Unmarshal(data, &server); err != nil {
		return fmt.Errorf("failed to parse server YAML: %w", err)
	}

	if err := server.NormalizeParent(); err != nil {
		return fmt.Errorf("invalid server configuration: %w", err)
	}

	if err := server.Validate(); err != nil {
		return fmt.Errorf("invalid server configuration: %w", err)
	}

	return CreateServer(server, "", false)
}

func init() {
	CreateServersCmd.Flags().String("address", "", "Server address (required)")
	CreateServersCmd.Flags().Int("port", 0, "Server port (required)")
	CreateServersCmd.Flags().Int("weight", defaultServerWeight, "Server weight (default: 100)")
	CreateServersCmd.Flags().Bool("ssl", false, "Enable SSL for the server")

	CreateServersCmd.Flags().StringP("output", "o", "", "Output format: yaml or json")
	CreateServersCmd.Flags().Bool("dry-run", false, "Simulate creation without actually applying")
}
