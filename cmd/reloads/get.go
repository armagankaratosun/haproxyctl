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

// Package reloads provides commands to inspect HAProxy reload history.
package reloads

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"haproxyctl/internal"

	"github.com/spf13/cobra"
)

// GetReloadsCmd represents "get reloads".
var GetReloadsCmd = &cobra.Command{
	Use:     "reloads [id]",
	Aliases: []string{"reload"},
	Short:   "List HAProxy reloads or fetch details of a specific reload",
	Long: `Retrieve HAProxy reload history from the Data Plane API.

Examples:
  haproxyctl get reloads
  haproxyctl get reloads 2019-01-03-44`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var id string
		if len(args) > 0 {
			id = args[0]
		}
		getReloads(cmd, id)
	},
}

// getReloads handles fetching reloads (list or single item).
func getReloads(cmd *cobra.Command, id string) {
	outputFormat := internal.GetFlagString(cmd, "output")
	if outputFormat == "" {
		outputFormat = "table"
	}

	var data interface{}
	var err error

	if id == "" {
		data, err = getReloadsListFromAPI(cmd)
	} else {
		data, err = getReloadFromAPI(cmd, id)
	}

	if err != nil {
		if id != "" && internal.IsNotFoundError(err) {
			_, _ = fmt.Fprintln(os.Stdout, internal.ResourceID("Reload", id)+" not found")
			return
		}
		log.Fatalf("Failed to fetch reload(s): %v", err)
	}

	internal.FormatOutput(data, outputFormat)
}

// getReloadsListFromAPI fetches the list of reloads.
func getReloadsListFromAPI(cmd *cobra.Command) ([]map[string]interface{}, error) {
	data, err := internal.SendRequestWithContext(cmd.Context(), "GET", "/services/haproxy/reloads", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reloads: %w", err)
	}

	var list []map[string]interface{}
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("failed to parse reloads list response: %w", err)
	}

	// Ensure deterministic ordering by reload ID when present.
	internal.SortByStringField(list, "id")

	return list, nil
}

// getReloadFromAPI fetches a single reload by id.
func getReloadFromAPI(cmd *cobra.Command, id string) (map[string]interface{}, error) {
	data, err := internal.SendRequestWithContext(cmd.Context(), "GET", "/services/haproxy/reloads/"+id, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reload %q: %w", id, err)
	}

	var r map[string]interface{}
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("failed to parse reload response: %w", err)
	}

	return r, nil
}

func init() {
	GetReloadsCmd.Flags().StringP("output", "o", "", "Output format: table, yaml, or json")
}
