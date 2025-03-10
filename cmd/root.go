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
	Short: "CLI tool for managing HAProxy",
	Long: `haproxyctl is a command-line tool for managing HAProxy backends, servers, frontends, and configuration.

Examples:
  haproxyctl get backends -o yaml
  haproxyctl create backend mybackend --mode http
  haproxyctl delete server mybackend myserver
  haproxyctl describe frontend myfrontend
  haproxyctl create -f backend.yaml
  haproxyctl create -f server.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		// Show help if no subcommands are provided
		fmt.Println("No command specified. Showing help:")
		cmd.Help()
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
