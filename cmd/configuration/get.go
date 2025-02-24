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
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

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
	Short: "Fetch HAProxy configuration version",
	Long:  `Retrieves the current HAProxy configuration version and returns it as JSON.`,
	Run: func(cmd *cobra.Command, args []string) {
		getConfigurationVersion()
	},
}

// getConfigurationRawCmd fetches the raw HAProxy configuration
var getConfigurationRawCmd = &cobra.Command{
	Use:   "raw",
	Short: "Fetch raw HAProxy configuration",
	Long:  `Retrieves the full raw HAProxy configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		getConfiguration()
	},
}

// getConfigurationVersion fetches the HAProxy configuration version and returns JSON output in a pretty format
func getConfigurationVersion() {
	data, err := utils.SendRequest("GET", "/services/haproxy/configuration/version", nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	versionStr := strings.TrimSpace(string(data)) // Remove newline and spaces

	versionInt, err := strconv.Atoi(versionStr)
	if err != nil {
		log.Fatal("Failed to parse version as an integer:", err)
	}

	// Create JSON output
	jsonOutput, err := json.MarshalIndent(map[string]int{"version": versionInt}, "", "    ")
	if err != nil {
		log.Fatal("Failed to format JSON:", err)
	}

	fmt.Println(string(jsonOutput)) // Print pretty JSON output
}

// getConfiguration fetches the raw HAProxy configuration
func getConfiguration() {
	data, err := utils.SendRequest("GET", "/services/haproxy/configuration/raw", nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data)) // Print raw HAProxy configuration
}

func init() {
	// Attach subcommands
	GetConfigurationCmd.AddCommand(getConfigurationVersionCmd)
	GetConfigurationCmd.AddCommand(getConfigurationRawCmd)
}
