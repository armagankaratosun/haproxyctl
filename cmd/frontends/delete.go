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
	"haproxyctl/internal"
	"log"
	"strconv"

	"github.com/spf13/cobra"
)

// DeleteFrontendsCmd represents "delete frontends"
var DeleteFrontendsCmd = &cobra.Command{
	Use:   "frontends <frontend_name>",
	Short: "Delete a specific HAProxy frontend",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		frontendName := args[0]
		deleteFrontend(frontendName)
	},
}

// deleteFrontend handles frontend deletion
func deleteFrontend(frontendName string) {
	version, err := internal.GetConfigurationVersion()
	if err != nil {
		log.Fatalf("Failed to fetch HAProxy configuration version: %v", err)
	}

	_, err = internal.SendRequest("DELETE",
		fmt.Sprintf("/services/haproxy/configuration/frontends/%s", frontendName),
		map[string]string{"version": strconv.Itoa(version)},
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to delete frontend '%s': %v", frontendName, err)
	}

	fmt.Printf("Frontend '%s' deleted successfully.\n", frontendName)
}
