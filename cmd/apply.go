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
	"errors"
	"fmt"
	"haproxyctl/cmd/backends"
	"haproxyctl/cmd/configuration"
	"haproxyctl/cmd/frontends"
	"haproxyctl/cmd/servers"
	"haproxyctl/internal"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var applyFile string

const (
	kindBackend  = "backend"
	kindFrontend = "frontend"
	kindServer   = "server"
	kindGlobal   = "global"
	kindDefaults = "defaults"
)

// applyCmd represents the top-level "apply" command.
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a manifest to HAProxy (create or replace)",
	Long: `Apply a manifest file in a declarative way, similar to kubectl apply.

The manifest must include apiVersion: haproxyctl/v1 and a supported kind
(Backend, Frontend, or Server). If the resource does not exist it will be
created; if it exists it will be replaced using the same logic as the
interactive edit flows.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if applyFile == "" {
			return errors.New("apply requires -f/--file")
		}
		return applyFromFile(cmd, applyFile)
	},
}

func applyFromFile(cmd *cobra.Command, filepath string) error {
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

	if metadata.APIVersion != "haproxyctl/v1" {
		return fmt.Errorf("unsupported apiVersion %q (expected haproxyctl/v1)", metadata.APIVersion)
	}

	outputFormat := cmd.Flags().Lookup("output").Value.String()
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	switch strings.ToLower(metadata.Kind) {
	case kindBackend:
		return backends.ApplyBackendFromYAML(data, outputFormat, dryRun)
	case kindFrontend:
		return frontends.ApplyFrontendFromYAML(data, outputFormat, dryRun)
	case kindGlobal:
		return configuration.ApplyGlobalFromYAML(data, outputFormat, dryRun)
	case kindDefaults:
		return configuration.ApplyDefaultsFromYAML(data, outputFormat, dryRun)
	case kindServer:
		var s servers.ServerConfig
		if err := yaml.Unmarshal(data, &s); err != nil {
			return fmt.Errorf("failed to parse server manifest: %w", err)
		}

		// For preview/dry-run, reuse the existing CreateServer behaviour to
		// render either the manifest or the payload without making changes.
		if outputFormat != "" || dryRun {
			return servers.CreateServer(s, outputFormat, dryRun)
		}

		if err := s.NormalizeParent(); err != nil {
			return fmt.Errorf("invalid server configuration: %w", err)
		}

		// Try to detect if the server exists.
		backendName := s.Parent
		endpoint := fmt.Sprintf("/services/haproxy/configuration/backends/%s/servers/%s", backendName, s.Name)
		_, err := internal.GetResource(endpoint)
		exists := err == nil

		if !exists && !internal.IsNotFoundError(err) {
			return fmt.Errorf("failed to check server existence: %w", err)
		}

		if exists {
			return servers.UpdateServer(s)
		}

		return servers.CreateServer(s, "", false)
	default:
		return fmt.Errorf("unsupported resource kind: %s (supported: Backend, Frontend, Server)", metadata.Kind)
	}
}

func init() {
	rootCmd.AddCommand(applyCmd)

	applyCmd.Flags().StringVarP(&applyFile, "file", "f", "", "Apply resource from YAML file (kind: Backend, Frontend, or Server)")
	applyCmd.Flags().StringP("output", "o", "", "Output format: yaml or json")
	applyCmd.Flags().Bool("dry-run", false, "Simulate apply without actually making changes")
}
