// cmd/frontends/create.go

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
	"log"
	"strconv"

	"haproxyctl/utils"

	"github.com/spf13/cobra"
)

// CreateFrontendsCmd represents "create frontends"
var CreateFrontendsCmd = &cobra.Command{
	Use:   "frontends <frontend_name>",
	Short: "Create a new HAProxy frontend (and optional binds)",
	Long: `Create a new HAProxy frontend either from a YAML file or CLI flags.

Examples:

  # imperatively, with a bind:
  haproxyctl create frontends myfront \
    --mode http \
    --default-backend webapp \
    --bind address=0.0.0.0,port=80,ssl=enabled

  # from manifest (no name on the command line):
  haproxyctl create frontends -f examples/frontend-with-binds.yaml`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var frontend frontendWithBinds

		// 1) Load from file if requested
		if fn := utils.GetFlagString(cmd, "file"); fn != "" {
			if err := frontend.LoadFromFile(fn); err != nil {
				log.Fatalf("failed to load frontend from file: %v", err)
			}
		} else {
			// 2) Otherwise require exactly one arg
			if len(args) != 1 {
				log.Fatalf("frontend name is required when not using -f")
			}
			frontend.LoadFromFlags(cmd, args[0])
		}

		// (rest of your existing logic follows…)
		if err := frontend.Validate(); err != nil {
			log.Fatalf("invalid frontend configuration: %v", err)
		}

		outFmt := utils.GetFlagString(cmd, "output")
		dryRun := utils.GetFlagBool(cmd, "dry-run")
		if outFmt != "" || dryRun {
			utils.FormatOutput(frontend, outFmt)
			if dryRun {
				fmt.Println("dry‑run mode enabled; no changes made.")
			}
			return
		}

		version, err := utils.GetConfigurationVersion()
		if err != nil {
			log.Fatalf("failed to fetch HAProxy version: %v", err)
		}
		apiObj := frontend.ToFrontendConfig()
		_, err = utils.SendRequest("POST",
			"/services/haproxy/configuration/frontends",
			map[string]string{"version": strconv.Itoa(version)},
			apiObj,
		)
		if err != nil {
			log.Fatalf("failed to create frontend %q: %v", apiObj.Name, err)
		}
		fmt.Printf("frontend %q created\n", apiObj.Name)

		for _, b := range frontend.Binds {
			if err := createBind(apiObj.Name, b); err != nil {
				log.Fatalf("failed to add bind to %q: %v", apiObj.Name, err)
			}
		}
	},
}

func init() {
	CreateFrontendsCmd.Flags().StringP("file", "f", "", "Load frontend config from YAML file")
	CreateFrontendsCmd.Flags().String("mode", "http", "Frontend mode (default: http)")
	CreateFrontendsCmd.Flags().String("default-backend", "", "Name of default backend")
	CreateFrontendsCmd.Flags().StringToString("forwardfor", nil, "ForwardFor settings (key=value)")
	CreateFrontendsCmd.Flags().String("timeout-client", "", "timeout client (e.g. 30s)")
	CreateFrontendsCmd.Flags().String("timeout-http-request", "", "timeout http request")
	CreateFrontendsCmd.Flags().String("timeout-http-keep-alive", "", "timeout http keep-alive")
	CreateFrontendsCmd.Flags().String("timeout-queue", "", "timeout queue")
	CreateFrontendsCmd.Flags().String("timeout-server", "", "timeout server")

	// Bind flag supports multiple binds
	CreateFrontendsCmd.Flags().StringArray("bind", nil,
		"Bind parameters (address=...,port=...,ssl=...). Repeat for multiple binds.")

	CreateFrontendsCmd.Flags().StringP("output", "o", "", "Output format: yaml or json")
	CreateFrontendsCmd.Flags().Bool("dry-run", false, "Simulate without applying")
}

// createBind POSTS a single BindConfig to an existing frontend.
func createBind(frontendName string, bind BindConfig) error {
	version, err := utils.GetConfigurationVersion()
	if err != nil {
		return fmt.Errorf("fetch version: %w", err)
	}
	endpoint := fmt.Sprintf("/services/haproxy/configuration/frontends/%s/binds", frontendName)
	_, err = utils.SendRequest("POST",
		endpoint,
		map[string]string{"version": strconv.Itoa(version)},
		bind,
	)
	return err
}
