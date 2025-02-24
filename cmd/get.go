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
	"haproxyctl/cmd/acls"
	"haproxyctl/cmd/backends"
	"haproxyctl/cmd/configuration"
	"haproxyctl/cmd/frontends"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve information from HAProxy",
	Long: `Fetch details about HAProxy configuration, including backends, frontends, and ACLs.
	
Examples:
  haproxyctl get configuration version -o json
  haproxyctl get backends -o yaml
  haproxyctl get frontends -o json
  haproxyctl get acls myfrontend -o yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("get called")
	},
}

func init() {
	// Add subcommands
	getCmd.AddCommand(configuration.GetConfigurationCmd)
	getCmd.AddCommand(backends.GetBackendsCmd)
	getCmd.AddCommand(frontends.GetFrontendsCmd)
	getCmd.AddCommand(acls.GetACLsCmd)

	// Add global output flag (all subcommands will inherit this)
	getCmd.PersistentFlags().StringP("output", "o", "", "Output format: yaml or json (default: table)")

	// Register the get command under root
	rootCmd.AddCommand(getCmd)
}
