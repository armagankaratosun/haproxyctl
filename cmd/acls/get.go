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
package acls

import (
	"fmt"
	"haproxyctl/internal"
	"log"

	"github.com/spf13/cobra"
)

// GetACLsCmd represents "get acls <frontend_name>"
var GetACLsCmd = &cobra.Command{
	Use:   "acls <frontend_name>",
	Short: "Retrieve ACLs for a specific HAProxy frontend",
	Args:  cobra.ExactArgs(1), // Requires exactly 1 argument: the frontend name
	Run: func(cmd *cobra.Command, args []string) {
		frontendName := args[0]
		getACLs(frontendName, cmd)
	},
}

// getACLs fetches the list of ACLs for a specific HAProxy frontend
func getACLs(frontendName string, cmd *cobra.Command) {
	endpoint := fmt.Sprintf("/services/haproxy/configuration/frontends/%s/acls", frontendName)

	data, err := internal.SendRequest("GET", endpoint, nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	outputFormat, _ := cmd.Flags().GetString("output")

	// Use FormatOutput to pretty-print JSON or YAML
	internal.FormatOutput(string(data), outputFormat)
}

func init() {
	// Ensure this command also inherits the `-o` flag
	GetACLsCmd.Flags().StringP("output", "o", "", "Output format: yaml or json")
}
