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
package frontends

import (
	"fmt"
	"haproxyctl/utils"
	"log"

	"github.com/spf13/cobra"
)

// GetFrontendsCmd represents "get frontends [frontend_name]"
var GetFrontendsCmd = &cobra.Command{
	Use:   "frontends [frontend_name]",
	Short: "Retrieve HAProxy frontends",
	Args:  cobra.MaximumNArgs(1), // Allows an optional frontend name
	Run: func(cmd *cobra.Command, args []string) {
		var frontendName string
		if len(args) > 0 {
			frontendName = args[0]
		}
		getFrontends(frontendName, cmd)
	},
}

// getFrontends fetches either all HAProxy frontends or a specific frontend if a name is provided
func getFrontends(frontendName string, cmd *cobra.Command) {
	var endpoint string
	if frontendName == "" {
		endpoint = "/services/haproxy/configuration/frontends"
	} else {
		endpoint = fmt.Sprintf("/services/haproxy/configuration/frontends/%s", frontendName)
	}

	data, err := utils.SendRequest("GET", endpoint, nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	outputFormat, _ := cmd.Flags().GetString("output")

	// Use FormatOutput to pretty-print JSON or YAML
	utils.FormatOutput(string(data), outputFormat)
}

func init() {
	// Ensure this command also inherits the `-o` flag
	GetFrontendsCmd.Flags().StringP("output", "o", "", "Output format: yaml or json")
}
