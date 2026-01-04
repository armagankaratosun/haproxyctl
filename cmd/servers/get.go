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

// Package servers provides commands to manage HAProxy backend servers.
package servers

import (
	"encoding/json"
	"fmt"
	"haproxyctl/internal"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// GetServersCmd represents "get servers".
var GetServersCmd = &cobra.Command{
	Use:     "servers <backend_name> [server_name]",
	Aliases: []string{"server"},
	Short:   "List servers in a backend, or fetch details for a specific server",
	Long: `Retrieve the list of servers in a backend, or details of a single server.

Examples:
  haproxyctl get servers mybackend
  haproxyctl get servers mybackend myserver`,
	Args: cobra.RangeArgs(1, serverArgsTwo),
	Run: func(cmd *cobra.Command, args []string) {
		backendName := args[0]
		var serverName string
		if len(args) > 1 {
			serverName = args[1]
		}
		getServers(cmd, backendName, serverName)
	},
}

// getServers fetches the list of servers or a specific server from a backend.
func getServers(cmd *cobra.Command, backendName, serverName string) {
	// First, ensure the backend exists so that a non-existent backend
	// does not quietly appear as "No resources found" when listing
	// servers.
	if _, err := internal.GetResource("/services/haproxy/configuration/backends/" + backendName); err != nil {
		if internal.IsNotFoundError(err) {
			_, _ = fmt.Fprintf(os.Stderr, "Error: backend %q not found\n\n", backendName)
			_ = cmd.Usage()
			return
		}
		log.Fatalf("Failed to fetch backend '%s': %v", backendName, err)
	}

	endpoint := fmt.Sprintf("/services/haproxy/configuration/backends/%s/servers", backendName)
	if serverName != "" {
		endpoint += "/" + serverName
	}
	data, err := internal.SendRequestWithContext(cmd.Context(), "GET", endpoint, nil, nil)
	if err != nil {
		if internal.IsNotFoundError(err) && serverName != "" {
			displayName := fmt.Sprintf("%s/%s", backendName, serverName)
			_, _ = fmt.Fprintln(os.Stdout, internal.ResourceID("Server", displayName)+" not found")
			return
		}
		log.Fatalf("Failed to fetch server(s) from backend '%s': %v", backendName, err)
	}

	format := internal.GetFlagString(cmd, "output")

	// Decode the JSON into a structured value so FormatOutput can
	// render tables / yaml / json consistently.
	var out interface{}
	if serverName == "" {
		var list []map[string]interface{}
		if err := json.Unmarshal(data, &list); err != nil {
			log.Fatalf("Failed to parse servers list response: %v\nResponse: %s", err, string(data))
		}

		// Ensure stable, predictable ordering of servers by name.
		internal.SortByStringField(list, "name")

		// For YAML/JSON, return a manifest-style List of Servers.
		// For table output, keep the existing flat list.
		if format == internal.OutputFormatYAML || format == "json" {
			items := make([]interface{}, 0, len(list))
			for _, srv := range list {
				items = append(items, mapServerResourceToConfig(backendName, srv))
			}
			out = internal.ManifestList{
				APIVersion: "haproxyctl/v1",
				Kind:       "List",
				Items:      items,
			}
		} else {
			var rows []interface{}
			for _, m := range list {
				rows = append(rows, m)
			}
			out = rows
		}
	} else {
		var srv map[string]interface{}
		if err := json.Unmarshal(data, &srv); err != nil {
			log.Fatalf("Failed to parse server response: %v\nResponse: %s", err, string(data))
		}

		if format == internal.OutputFormatYAML || format == "json" {
			out = mapServerResourceToConfig(backendName, srv)
		} else {
			out = srv
		}
	}

	internal.FormatOutput(out, format)
}

// mapServerResourceToConfig converts a raw API server object into a
// ServerConfig suitable for manifest-style output (e.g. get -o yaml).
func mapServerResourceToConfig(backendName string, obj map[string]interface{}) ServerConfig {
	var sc ServerConfig

	sc.APIVersion = "haproxyctl/v1"
	sc.Kind = "Server"

	if v, ok := obj["name"].(string); ok {
		sc.Name = v
	}
	if v, ok := obj["address"].(string); ok {
		sc.Address = v
	}
	if p, ok := obj["port"].(float64); ok {
		sc.Port = int(p)
	}
	if w, ok := obj["weight"].(float64); ok {
		sc.Weight = int(w)
	}
	if v, ok := obj["ssl"].(string); ok && v == "enabled" {
		sc.SSL = true
	}

	sc.Backend = backendName
	return sc
}

func init() {
	// Inherit the global -o flag for output formatting
	GetServersCmd.Flags().StringP("output", "o", "", "Output format: yaml or json")
}
