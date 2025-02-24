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
	"haproxyctl/utils"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

// GetBackendsCmd represents "get backends"
var GetBackendsCmd = &cobra.Command{
	Use:   "backends [backend_name]",
	Short: "Retrieve HAProxy backends",
	Args:  cobra.MaximumNArgs(1), // Allow optional backend name
	Run: func(cmd *cobra.Command, args []string) {
		var backendName string
		if len(args) > 0 {
			backendName = args[0]
		}
		getBackends(backendName, cmd)
	},
}

// getBackends fetches either all HAProxy backends or a specific backend if a name is provided
func getBackends(backendName string, cmd *cobra.Command) {
	var endpoint string
	if backendName == "" {
		endpoint = "/services/haproxy/configuration/backends"
	} else {
		endpoint = fmt.Sprintf("/services/haproxy/configuration/backends/%s", backendName)
	}

	data, err := utils.SendRequest("GET", endpoint, nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	outputFormat, _ := cmd.Flags().GetString("output")

	// Use FormatOutput globally
	utils.FormatOutput(strings.TrimSpace(string(data)), outputFormat)
}
