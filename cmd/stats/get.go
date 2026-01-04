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

// Package stats provides commands to inspect HAProxy runtime statistics.
package stats

import (
	"encoding/json"
	"fmt"
	"os"

	"haproxyctl/internal"

	"github.com/spf13/cobra"
)

// GetStatsCmd represents "get stats".
var GetStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Retrieve HAProxy runtime statistics",
	Long: `Fetch HAProxy runtime statistics from the Data Plane API.

By default, results are printed as JSON, since the native stats structure
contains many fields and is best consumed programmatically.

Examples:
  haproxyctl get stats -o json
  haproxyctl get stats --type backend --name s3_backend -o json
  haproxyctl get stats --type server --parent s3_backend --name archive`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		getStats(cmd)
	},
}

// getStats handles fetching native stats with optional filters.
func getStats(cmd *cobra.Command) {
	outputFormat := internal.GetFlagString(cmd, "output")
	if outputFormat == "" {
		outputFormat = "json"
	}

	objType := internal.GetFlagString(cmd, "type")
	name := internal.GetFlagString(cmd, "name")
	parent := internal.GetFlagString(cmd, "parent")

	data, err := getNativeStatsFromAPI(cmd, objType, name, parent)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to fetch HAProxy stats: %v\n", err)
		return
	}

	internal.FormatOutput(data, outputFormat)
}

// getNativeStatsFromAPI calls the Data Plane API /stats/native endpoint.
func getNativeStatsFromAPI(cmd *cobra.Command, objType, name, parent string) (map[string]interface{}, error) {
	query := map[string]string{}
	if objType != "" {
		query["type"] = objType
	}
	if name != "" {
		query["name"] = name
	}
	if parent != "" {
		query["parent"] = parent
	}

	raw, err := internal.SendRequestWithContext(cmd.Context(), "GET", "/services/haproxy/stats/native", query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch native stats: %w", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse native stats response: %w", err)
	}

	return payload, nil
}

func init() {
	GetStatsCmd.Flags().String("type", "", "Object type to get stats for: frontend, backend, or server")
	GetStatsCmd.Flags().String("name", "", "Object name to get stats for")
	GetStatsCmd.Flags().String("parent", "", "Parent name (for server stats, the backend name)")
}
