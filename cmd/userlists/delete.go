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

// Package userlists provides commands to manage HAProxy userlists, users, and groups.
package userlists

import (
	"fmt"
	"haproxyctl/internal"
	"log"
	"strconv"

	"github.com/spf13/cobra"
)

// DeleteUserlistsCmd represents "delete userlists".
var DeleteUserlistsCmd = &cobra.Command{
	Use:   "userlists <name>",
	Short: "Delete a HAProxy userlist",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		name := args[0]
		if err := DeleteUserlistByName(name); err != nil {
			log.Fatalf("Failed to delete userlist %q: %v", name, err)
		}
	},
}

// DeleteUserlistByName deletes a userlist using the configuration versioned DELETE.
func DeleteUserlistByName(name string) error {
	version, err := internal.GetConfigurationVersion()
	if err != nil {
		return fmt.Errorf("failed to fetch HAProxy configuration version: %w", err)
	}

	endpoint := "/services/haproxy/configuration/userlists/" + name
	_, err = internal.SendRequest(
		"DELETE",
		endpoint,
		map[string]string{"version": strconv.Itoa(version)},
		nil,
	)
	if err != nil {
		return internal.FormatAPIError("Userlist", name, "delete", err)
	}

	internal.PrintStatus("Userlist", name, internal.ActionDeleted)
	return nil
}
