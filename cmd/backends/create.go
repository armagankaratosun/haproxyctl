// Package backends provides commands to manage HAProxy backends.
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
package backends

import (
	"fmt"
	"haproxyctl/cmd/servers"
	"haproxyctl/internal"
	"log"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// CreateBackendsCmd represents "create backends".
var CreateBackendsCmd = &cobra.Command{
	Use:   "backends <backend_name>",
	Short: "Create a new HAProxy backend",
	Long: `Create a new HAProxy backend either from a YAML file or CLI flags.

Examples:
  haproxyctl create backends mybackend \
    --mode http \
    --balance algorithm=roundrobin \
    --server name=s1,address=10.0.0.1,port=80,weight=100 \
    --server name=s2,address=10.0.0.2,port=8080,weight=200
  haproxyctl create backends mybackend -f mybackend.yaml`,

	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		backendName := args[0]

		var backendWithServers backendWithServers
		backendWithServers.LoadFromFlags(cmd, backendName)

		if err := backendWithServers.Validate(); err != nil {
			log.Fatalf("Invalid backend configuration: %v", err)
		}

		outputFormat := internal.GetFlagString(cmd, "output")
		dryRun := internal.GetFlagBool(cmd, "dry-run")

		if err := createBackend(backendWithServers, outputFormat, dryRun); err != nil {
			log.Fatalf("Failed to create backend: %v", err)
		}
	},
}

// CreateBackendFromFile is used for "haproxyctl create -f file.yaml".
func CreateBackendFromFile(data []byte) error {
	var backendWithServers backendWithServers
	if err := yaml.Unmarshal(data, &backendWithServers); err != nil {
		return fmt.Errorf("failed to parse backend configuration file: %w", err)
	}

	if err := backendWithServers.Validate(); err != nil {
		return fmt.Errorf("invalid backend configuration: %w", err)
	}

	return createBackend(backendWithServers, "", false)
}

// createBackend handles backend creation with validation.
func createBackend(backendWithServers backendWithServers, outputFormat string, dryRun bool) error {
	if outputFormat != "" || dryRun {
		// For structured formats, preview the actual API payload
		// (timeouts in ms, tcpka/redispatch encoded). For the
		// default case (no -o), render a YAML view of the richer
		// backendWithServers structure.
		if outputFormat == "" {
			outputFormat = internal.OutputFormatYAML
			internal.FormatOutput(backendWithServers, outputFormat)
		} else {
			internal.FormatOutput(backendWithServers.toPayload(), outputFormat)
		}
		if dryRun {
			if _, err := fmt.Fprintln(os.Stdout, "Dry run mode enabled. No changes made."); err != nil {
				log.Printf("warning: failed to write dry‑run message: %v", err)
			}
		}
		return nil
	}

	version, err := internal.GetConfigurationVersion()
	if err != nil {
		return fmt.Errorf("failed to fetch HAProxy configuration version: %w", err)
	}

	_, err = internal.SendRequest("POST", "/services/haproxy/configuration/backends",
		map[string]string{"version": strconv.Itoa(version)},
		backendWithServers.toPayload(),
	)
	if err != nil {
		return fmt.Errorf("failed to create backend '%s': %w", backendWithServers.Name, err)
	}
	if _, err := fmt.Fprintf(os.Stdout, "Backend '%s' created successfully.\n", backendWithServers.Name); err != nil {
		log.Printf("warning: failed to write backend created message: %v", err)
	}

	// Create attached servers if any
	for _, server := range backendWithServers.Servers {
		server.Backend = backendWithServers.Name // Set the backend name directly
		if err := servers.CreateServer(server, outputFormat, dryRun); err != nil {
			return fmt.Errorf("failed to create server '%s' for backend '%s': %w", server.Name, backendWithServers.Name, err)
		}
	}

	return nil
}

func init() {
	CreateBackendsCmd.Flags().String("mode", "http", "Backend mode (default: http)")
	CreateBackendsCmd.Flags().StringToString("balance", map[string]string{"algorithm": "roundrobin"}, "Balance settings (key=value)")
	CreateBackendsCmd.Flags().StringToString("default-server", nil, "Default server settings (key=value)")
	CreateBackendsCmd.Flags().StringToString("forwardfor", nil, "ForwardFor settings (key=value)")

	CreateBackendsCmd.Flags().String("timeout-client", "", "Client timeout (e.g., 30s)")
	CreateBackendsCmd.Flags().String("timeout-queue", "", "Queue timeout (e.g., 30s)")
	CreateBackendsCmd.Flags().String("timeout-server", "", "Server timeout (e.g., 30s)")

	CreateBackendsCmd.Flags().Bool("redispatch", false, "Enable redispatch")

	// Server flag supports multiple servers
	CreateBackendsCmd.Flags().StringArray("server", nil, "Define server (name=s1,address=10.0.0.1,port=80,weight=100). Repeat for multiple servers.")

	// Output and dry-run
	CreateBackendsCmd.Flags().StringP("output", "o", "", "Output format: yaml or json")
	CreateBackendsCmd.Flags().Bool("dry-run", false, "Simulate creation without actually applying")
}
