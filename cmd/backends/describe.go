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
	"haproxyctl/internal"
	"log"

	"github.com/spf13/cobra"
)

// DescribeBackendsCmd represents "describe backends"
var DescribeBackendsCmd = &cobra.Command{
	Use:   "backends <backend_name>",
	Short: "Describe a specific HAProxy backend and its servers",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		backendName := args[0]
		describeBackend(backendName)
	},
}

// describeBackend fetches a backend and its servers, and prints a detailed description
func describeBackend(backendName string) {
	backend, err := internal.GetResource(fmt.Sprintf("/services/haproxy/configuration/backends/%s", backendName))
	if err != nil {
		log.Fatalf("Failed to fetch backend '%s': %v", backendName, err)
	}

	servers, err := internal.GetResourceList(fmt.Sprintf("/services/haproxy/configuration/backends/%s/servers", backendName))
	if err != nil {
		log.Fatalf("Failed to fetch servers for backend '%s': %v", backendName, err)
	}

	internal.PrintResourceDescription("Backend", backend, backendDescriptionSections(), servers)
}

// backendDescriptionSections defines the sections and fields to display in backend descriptions
func backendDescriptionSections() map[string][]string {
	return map[string][]string{
		"basic":          {"name", "mode", "balance"},
		"timeouts":       {"timeout_client", "timeout_queue", "timeout_server"},
		"advanced":       {"tcpka", "redispatch"},
		"default_server": {"alpn", "check", "check_alpn", "maxconn", "weight"},
	}
}

func init() {
	DescribeBackendsCmd.Flags().StringP("output", "o", "", "Output format: table, yaml, or json")
}
