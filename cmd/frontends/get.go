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
package frontends

import (
	"fmt"
	"haproxyctl/utils"
	"log"

	"github.com/spf13/cobra"
)

// GetFrontendsCmd represents "get frontends"
var GetFrontendsCmd = &cobra.Command{
	Use:   "frontends [frontend_name]",
	Short: "Retrieve HAProxy frontends",
	Args:  cobra.MaximumNArgs(1), // Allows an optional frontend name
	Run: func(cmd *cobra.Command, args []string) {
		var frontendName string
		if len(args) > 0 {
			frontendName = args[0]
		}
		getFrontends(cmd, frontendName)
	},
}

func getFrontends(cmd *cobra.Command, frontendName string) {
	var data interface{}
	var err error

	if frontendName != "" {
		data, err = utils.GetResource(fmt.Sprintf("/services/haproxy/configuration/frontends/%s", frontendName))
	} else {
		data, err = utils.GetResourceList("/services/haproxy/configuration/frontends")
	}

	if err != nil {
		log.Fatalf("Failed to fetch frontend(s): %v", err)
	}

	outputFormat := utils.GetFlagString(cmd, "output")

	if outputFormat == "" {
		outputFormat = "table"
	}

	utils.FormatOutput(data, outputFormat)
}

func init() {
	// Ensure this command also inherits the `-o` flag
	GetFrontendsCmd.Flags().StringP("output", "o", "", "Output format: yaml or json")
}
