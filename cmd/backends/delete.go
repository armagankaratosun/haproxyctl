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
	"haproxyctl/internal"
	"log"
	"strconv"

	"github.com/spf13/cobra"
)

// DeleteBackendsCmd represents "delete backends".
var DeleteBackendsCmd = &cobra.Command{
	Use:   "backends <backend_name>",
	Short: "Delete a specific HAProxy backend",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		backendName := args[0]
		deleteBackend(backendName)
	},
}

// deleteBackend handles backend deletion.
func deleteBackend(backendName string) {
	version, err := internal.GetConfigurationVersion()
	if err != nil {
		log.Fatalf("Failed to fetch HAProxy configuration version: %v", err)
	}

	endpoint := "/services/haproxy/configuration/backends/" + backendName
	_, err = internal.SendRequest("DELETE",
		endpoint,
		map[string]string{"version": strconv.Itoa(version)},
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to delete backend '%s': %v", backendName, err)
	}

	internal.PrintStatus("Backend", backendName, internal.ActionDeleted)
}
