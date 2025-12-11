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

// Package configuration provides commands to manage HAProxy global configuration.
package configuration

import (
	"fmt"
	"log"
	"os"

	"haproxyctl/internal"

	"github.com/spf13/cobra"
)

const outputFormatJSON = "json"
const globalRawHint = "configuration/globals no rules defined; use 'haproxyctl get configuration raw' and 'haproxyctl create configuration raw' for global settings"

// GetConfigurationCmd represents the "get configuration" command.
var GetConfigurationCmd = &cobra.Command{
	Use:   "configuration",
	Short: "Fetch HAProxy configuration",
	Long:  `Retrieve details about HAProxy configuration, including the version and raw configuration.`,
}

// getConfigurationGlobalCmd fetches the HAProxy globals section via JSON.
var getConfigurationGlobalCmd = &cobra.Command{
	Use:     "globals",
	Aliases: []string{"global"},
	Short:   "Retrieves HAProxy global configuration",
	Long:    `Retrieves the HAProxy "globals" section as JSON/YAML.`,
	Run: func(cmd *cobra.Command, _ []string) {
		outputFormat := internal.GetFlagString(cmd, "output")

		obj, err := internal.GetResource("/services/haproxy/configuration/global")
		if err != nil {
			log.Fatalf("Failed to fetch global configuration: %v", err)
		}

		cfg := mapGlobalFromAPI(obj)

		// If the API returns an empty JSON object for globals, there is no
		// structured representation available; direct users to the raw config.
		if outputFormat == "" && cfg.isEmpty() {
			_, _ = fmt.Fprintln(os.Stdout, globalRawHint)
			return
		}

		if outputFormat == "" {
			row := map[string]interface{}{
				"daemon":        cfg.Daemon,
				"nbproc":        cfg.Nbproc,
				"maxconn":       cfg.Maxconn,
				"log":           cfg.Log,
				"log_send_host": cfg.LogSendHost,
				"stats_socket":  cfg.StatsSocket,
				"stats_timeout": cfg.StatsTimeout,
				"spread_checks": cfg.SpreadChecks,
			}
			internal.FormatOutput(row, "")
			return
		}

		internal.FormatOutput(cfg, outputFormat)
	},
}

// getConfigurationDefaultsCmd fetches the HAProxy defaults section via JSON.
var getConfigurationDefaultsCmd = &cobra.Command{
	Use:   "defaults",
	Short: "Retrieves HAProxy defaults configuration",
	Long:  `Retrieves the HAProxy "defaults" section as JSON/YAML.`,
	Run: func(cmd *cobra.Command, _ []string) {
		outputFormat := internal.GetFlagString(cmd, "output")

		list, err := internal.GetResourceList("/services/haproxy/configuration/defaults")
		if err != nil {
			if internal.IsNotFoundError(err) {
				_, _ = fmt.Fprintln(os.Stdout, "configuration/defaults no rules defined")
				return
			}
			log.Fatalf("Failed to fetch defaults configuration: %v", err)
		}

		if len(list) == 0 {
			_, _ = fmt.Fprintln(os.Stdout, "configuration/defaults no rules defined")
			return
		}

		cfg := mapDefaultsFromAPI(list[0])

		if outputFormat == "" && cfg.isEmpty() {
			_, _ = fmt.Fprintln(os.Stdout, "configuration/defaults no rules defined")
			return
		}

		if outputFormat == "" {
			row := map[string]interface{}{
				"name":            cfg.Name,
				"mode":            cfg.Mode,
				"timeout_client":  cfg.TimeoutClient,
				"timeout_server":  cfg.TimeoutServer,
				"timeout_connect": cfg.TimeoutConnect,
				"timeout_queue":   cfg.TimeoutQueue,
				"timeout_tunnel":  cfg.TimeoutTunnel,
				"balance":         cfg.Balance,
				"log":             cfg.Log,
			}
			internal.FormatOutput(row, "")
			return
		}

		internal.FormatOutput(cfg, outputFormat)
	},
}

// getConfigurationVersionCmd fetches the HAProxy configuration version.
var getConfigurationVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Retrieves HAProxy configuration version in JSON format",
	Long:  `Retrieves the current HAProxy configuration version.`,
	Run: func(cmd *cobra.Command, _ []string) {
		GetConfigurationVersion(cmd)
	},
}

// getConfigurationRawCmd fetches the raw HAProxy configuration.
var getConfigurationRawCmd = &cobra.Command{
	Use:   "raw",
	Short: "Retrieves raw HAProxy configuration",
	Long:  `Retrieves the full raw HAProxy configuration.`,
	Run: func(cmd *cobra.Command, _ []string) {
		data, err := GetConfigurationRaw(cmd)
		if err != nil {
			log.Fatalf("Failed to fetch raw configuration: %v", err)
		}
		cmd.Println(string(data))
	},
}

// GetConfigurationVersion fetches and displays the configuration version.
func GetConfigurationVersion(cmd *cobra.Command) {
	outputFormat := internal.GetFlagString(cmd, "output")

	version, err := internal.GetConfigurationVersion()
	if err != nil {
		log.Fatalf("Failed to fetch HAProxy configuration version: %v", err)
	}

	// Build a structured object to support multiple output formats
	versionData := map[string]int{"version": version}

	// Default to JSON if no explicit format is set, I know this is an ugly hack
	if outputFormat == "" {
		outputFormat = outputFormatJSON
	}

	internal.FormatOutput(versionData, outputFormat)
}

// GetConfigurationRaw fetches the raw HAProxy configuration.
func GetConfigurationRaw(cmd *cobra.Command) ([]byte, error) {
	return internal.SendRequestWithContext(cmd.Context(), "GET", "/services/haproxy/configuration/raw", nil, nil)
}

func init() {
	// Attach subcommands.
	GetConfigurationCmd.AddCommand(getConfigurationVersionCmd)
	GetConfigurationCmd.AddCommand(getConfigurationRawCmd)
	GetConfigurationCmd.AddCommand(getConfigurationGlobalCmd)
	GetConfigurationCmd.AddCommand(getConfigurationDefaultsCmd)
}
