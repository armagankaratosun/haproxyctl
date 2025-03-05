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
	"haproxyctl/utils"
	"log"

	"github.com/spf13/cobra"
)

// GetBackendsCmd represents "get backends"
var GetBackendsCmd = &cobra.Command{
	Use:   "backends [backend_name]",
	Short: "List HAProxy backends or fetch details of a specific backend",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var backendName string
		if len(args) > 0 {
			backendName = args[0]
		}
		getBackends(cmd, backendName)
	},
}

// getBackends handles fetching backends (list or single item)
func getBackends(cmd *cobra.Command, backendName string) {
	outputFormat := utils.GetFlagString(cmd, "output")

	if outputFormat == "" {
		outputFormat = "table" // Default to table if not specified
	}

	var data interface{}
	var err error

	if backendName == "" {
		// Fetch all backends (list)
		data, err = utils.GetResourceList("/services/haproxy/configuration/backends")
		if err != nil {
			log.Fatalf("Failed to fetch backends: %v", err)
		}

		// Enrich each backend with servers (applies only to tables, but harmless for yaml/json)
		if backendList, ok := data.([]map[string]interface{}); ok {
			for i := range backendList {
				utils.EnrichBackendWithServers(backendList[i])
			}
		}
	} else {
		// Fetch a specific backend (single object)
		data, err = utils.GetResource("/services/haproxy/configuration/backends/" + backendName)
		if err != nil {
			log.Fatalf("Failed to fetch backend '%s': %v", backendName, err)
		}

		if backend, ok := data.(map[string]interface{}); ok {
			utils.EnrichBackendWithServers(backend)
		}
	}

	utils.FormatOutput(data, outputFormat)
}

func init() {
	GetBackendsCmd.Flags().StringP("output", "o", "", "Output format: table, yaml, or json")
}
