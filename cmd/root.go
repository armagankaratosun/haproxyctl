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
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "haproxyctl",
	Short: "CLI tool for managing HAProxy through Data Plane API v3",
	Long: `haproxyctl is a command-line tool for managing and creating HAProxy objects supported in the HAProxy Data Plane API v3.

Find more information at: https://www.haproxy.com/documentation/dataplaneapi/

Basic Commands:
  login           Create HAProxy Data Plane API configuration file
  get             Display one or more HAProxy resources (backends, servers, frontends)
  create          Create a new HAProxy resource from a file or CLI flags
  delete          Delete an existing HAProxy resource
  describe        Show details of a specific HAProxy resource

Resource-Specific Commands:
  backends        Manage HAProxy backends
  servers         Manage HAProxy servers within a backend
  frontends       Manage HAProxy frontends
  acls            Manage HAProxy ACLs

Usage:
  haproxyctl [command] [options]

Examples:
  # Retrieve all backends in YAML format
  haproxyctl get backends -o yaml

  # Create a backend with round-robin load balancing
  haproxyctl create backend mybackend --mode http --balance algorithm=roundrobin

  # Add a server to an existing backend
  haproxyctl create server mybackend myserver --address 10.0.0.1 --port 80 --weight 100

  # Describe a frontend with details
  haproxyctl describe frontend myfrontend

  # Delete a server from a backend
  haproxyctl delete server mybackend myserver

  # Create resources from YAML manifests
  haproxyctl create -f backend.yaml
  haproxyctl create -f server.yaml

Use "haproxyctl <command> --help" for more information about a given command.
`,

	Run: func(cmd *cobra.Command, args []string) {
		// Stool for managing HAProxy backends, sehow help if no subcommands are provided
		fmt.Println("No command specified. Showing help:")
		if err := cmd.Help(); err != nil {
			fmt.Fprintf(os.Stderr, "Error displaying help: %v\n", err)
		}
	},
}

// Execute runs the root command and dispatches subcommands
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Define global flags (if needed in the future)
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.haproxyctl.yaml)")

	// Ensure rootCmd shows help when run without arguments
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true}) // Hide default help command
}
