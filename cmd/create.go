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

// Package cmd wires top-level CLI commands for haproxyctl.
package cmd

import (
	"fmt"
	"haproxyctl/cmd/backends"
	"haproxyctl/cmd/certificates"
	"haproxyctl/cmd/configuration"
	"haproxyctl/cmd/frontends"
	"haproxyctl/cmd/servers"
	"haproxyctl/cmd/userlists"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var createFile string

// createCmd represents the top-level "create" command.
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a resource in HAProxy",
	RunE: func(_ *cobra.Command, _ []string) error {
		if createFile != "" {
			return createFromFile(createFile)
		}

		log.Fatal("Specify a resource type (backends, servers) and its name, or use '-f' to create from file.")
		return nil
	},
}

func createFromFile(filepath string) error {
	// The CLI is expected to read user-specified manifest files.
	data, err := os.ReadFile(filepath) //nolint:gosec // filepath comes from user input by design
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filepath, err)
	}

	// Extract kind and apiVersion to figure out what to do.
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
	case "userlist":
		return userlists.CreateUserlistFromFile(data)
	default:
		return fmt.Errorf("unsupported resource kind: %s (supported: Backend, Server, Userlist)", metadata.Kind)
	}
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Add subcommands for explicit CLI resource creation (with positional args).
	createCmd.AddCommand(backends.CreateBackendsCmd)
	createCmd.AddCommand(certificates.CreateCertificatesCmd)
	createCmd.AddCommand(servers.CreateServersCmd)
	createCmd.AddCommand(frontends.CreateFrontendsCmd)
	createCmd.AddCommand(configuration.CreateConfigurationCmd)

	// Global flag for file-based creation (works for multiple resource kinds)
	createCmd.Flags().StringVarP(&createFile, "file", "f", "", "Create resource from YAML file (supports kind: Backend, Server, and Userlist)")
}
