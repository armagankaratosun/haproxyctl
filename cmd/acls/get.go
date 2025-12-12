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

// Package acls provides commands to manage HAProxy ACLs via the Data Plane API.
package acls

import (
	"encoding/json"
	"fmt"
	"haproxyctl/internal"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// indirection for easier testing.
var getACLsRequest = internal.SendRequestWithContext

// GetACLsCmd represents "get acls <frontend_name>".
var GetACLsCmd = &cobra.Command{
	Use:     "acls <frontend_name>",
	Aliases: []string{"acl"},
	Short:   "Retrieve ACLs for a specific HAProxy frontend",
	Args:    cobra.ExactArgs(1), // Requires exactly 1 argument: the frontend name
	Run: func(cmd *cobra.Command, args []string) {
		frontendName := args[0]
		getACLs(frontendName, cmd)
	},
}

// getACLs fetches the list of ACLs for a specific HAProxy frontend.
func getACLs(frontendName string, cmd *cobra.Command) {
	endpoint := fmt.Sprintf("/services/haproxy/configuration/frontends/%s/acls", frontendName)

	data, err := getACLsRequest(cmd.Context(), "GET", endpoint, nil, nil)
	if err != nil {
		if internal.IsNotFoundError(err) {
			_, _ = fmt.Fprintf(os.Stderr, "Error: frontend %q not found\n\n", frontendName)
			_ = cmd.Usage()
			return
		}
		log.Fatal(err)
	}

	outputFormat, _ := cmd.Flags().GetString("output")

	var acls []map[string]interface{}
	if err := json.Unmarshal(data, &acls); err != nil {
		log.Fatalf("failed to parse ACL response: %v\nResponse: %s", err, string(data))
	}

	// Ensure deterministic ordering of ACLs by acl_name when listing.
	internal.SortByStringField(acls, "acl_name")

	// Use FormatOutput to pretty-print JSON or YAML
	internal.FormatOutput(acls, outputFormat)
}

func init() {
	// Ensure this command also inherits the `-o` flag
	GetACLsCmd.Flags().StringP("output", "o", "", "Output format: yaml or json")
}
