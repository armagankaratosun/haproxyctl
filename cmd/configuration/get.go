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
package configuration

import (
	"fmt"
	"log"

	"haproxyctl/utils"

	"github.com/spf13/cobra"
)

// GetConfigurationCmd represents the "get configuration" command
var GetConfigurationCmd = &cobra.Command{
	Use:   "configuration",
	Short: "Fetch HAProxy configuration",
	Long:  `Retrieve details about HAProxy configuration, including the version and raw configuration.`,
}

// getConfigurationVersionCmd fetches the HAProxy configuration version
var getConfigurationVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Retrieves HAProxy configuration version in JSON format",
	Long:  `Retrieves the current HAProxy configuration version.`,
	Run: func(cmd *cobra.Command, args []string) {
		GetConfigurationVersion(cmd)
	},
}

// getConfigurationRawCmd fetches the raw HAProxy configuration
var getConfigurationRawCmd = &cobra.Command{
	Use:   "raw",
	Short: "Retrieves raw HAProxy configuration",
	Long:  `Retrieves the full raw HAProxy configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		GetConfigurationRaw()
	},
}

// GetConfigurationVersion fetches and displays the configuration version
func GetConfigurationVersion(cmd *cobra.Command) {
	outputFormat := utils.GetFlagString(cmd, "output")

	version, err := utils.GetConfigurationVersion()
	if err != nil {
		log.Fatalf("Failed to fetch HAProxy configuration version: %v", err)
	}

	// Build a structured object to support multiple output formats
	versionData := map[string]int{"version": version}

	// Default to JSON if no explicit format is set, I know this is an ugly hack
	if outputFormat == "" {
		outputFormat = "json"
	}

	utils.FormatOutput(versionData, outputFormat)
}

// GetConfigurationRaw fetches and displays the raw HAProxy configuration
func GetConfigurationRaw() {
	data, err := utils.SendRequest("GET", "/services/haproxy/configuration/raw", nil, nil)
	if err != nil {
		log.Fatalf("Failed to fetch raw configuration: %v", err)
	}

	fmt.Println(string(data))
}

func init() {
	// Attach subcommands
	GetConfigurationCmd.AddCommand(getConfigurationVersionCmd)
	GetConfigurationCmd.AddCommand(getConfigurationRawCmd)
}
