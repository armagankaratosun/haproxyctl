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
	"haproxyctl/internal"
	"log"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

// CreateConfigurationCmd represents "create configuration"
var CreateConfigurationCmd = &cobra.Command{
	Use:   "configuration",
	Short: "Manage HAProxy configuration",
	Long: `Create or update HAProxy configuration via the Data Plane API.

You can push a raw HAProxy config file, automatically handling version checks.`,
}

// createConfigRawCmd represents "create configuration raw"
var createConfigRawCmd = &cobra.Command{
	Use:   "raw [file_path]",
	Short: "Upload a raw HAProxy configuration file",
	Long: `Upload a raw HAProxy configuration file to HAProxy via the Data Plane API.

You may specify the path either as a positional argument or with -f/--file.

Examples:
  haproxyctl create configuration raw /etc/haproxy/haproxy.cfg
  haproxyctl create configuration raw -f /etc/haproxy/haproxy.cfg`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// First, grab the shared "-f/--file" flag from the parent `create` command
		fileFlag, _ := cmd.Flags().GetString("file")

		var path string
		switch {
		case fileFlag != "" && len(args) > 0:
			log.Fatal("specify either a positional file or --file, not both")
		case fileFlag != "":
			path = fileFlag
		case len(args) == 1:
			path = args[0]
		default:
			log.Fatal("file path is required (positional or --file)")
		}

		// Read the raw HAProxy config
		data, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("failed to read %s: %v", path, err)
		}

		// Fetch the current HAProxy config version
		version, err := internal.GetConfigurationVersion()
		if err != nil {
			log.Fatalf("failed to fetch HAProxy configuration version: %v", err)
		}

		// POST the raw config
		endpoint := "/services/haproxy/configuration/raw"
		if _, err := internal.SendRawRequest(
			"POST",
			endpoint,
			map[string]string{"version": strconv.Itoa(version)},
			data,
			"text/plain",
		); err != nil {
			log.Fatalf("failed to push raw configuration: %v", err)
		}

		fmt.Println("raw configuration pushed")
	},
}

func init() {
	CreateConfigurationCmd.AddCommand(createConfigRawCmd)

}
