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
package cmd

import (
	"fmt"
	"haproxyctl/cmd/backends"
	"haproxyctl/cmd/configuration"
	"haproxyctl/cmd/frontends"
	"haproxyctl/cmd/servers"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var createFile string

// createCmd represents the top-level "create" command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a resource in HAProxy",
	RunE: func(cmd *cobra.Command, args []string) error {
		if createFile != "" {
			return createFromFile(createFile)
		}

		log.Fatal("Specify a resource type (backends, servers) and its name, or use '-f' to create from file.")
		return nil
	},
}

func createFromFile(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filepath, err)
	}

	// Extract kind and apiVersion to figure out what to do
	var metadata struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
	}
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return fmt.Errorf("failed to parse YAML file: %w", err)
	}

	kind := strings.ToLower(metadata.Kind)

	switch kind {
	case "backend":
		return backends.CreateBackendFromFile(data)
	case "server":
		return servers.CreateServerFromFile(data)
	default:
		return fmt.Errorf("unsupported resource kind: %s (supported: Backend, Server)", metadata.Kind)
	}
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Add subcommands for explicit CLI resource creation (with positional args)
	createCmd.AddCommand(backends.CreateBackendsCmd)
	createCmd.AddCommand(servers.CreateServersCmd)
	createCmd.AddCommand(frontends.CreateFrontendsCmd)
	createCmd.AddCommand(configuration.CreateConfigurationCmd)

	// Global flag for file-based creation (works for both backends and servers)
	createCmd.Flags().StringVarP(&createFile, "file", "f", "", "Create resource from YAML file (supports kind: Backend and kind: Server)")

}
