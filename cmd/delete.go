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
	"os"
	"strconv"
	"strings"

	"haproxyctl/cmd/backends"
	"haproxyctl/cmd/certificates"
	"haproxyctl/cmd/frontends"
	"haproxyctl/cmd/servers"
	"haproxyctl/cmd/userlists"
	"haproxyctl/internal"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var deleteFile string

// deleteCmd represents the delete command.
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete resources in HAProxy",
	RunE: func(_ *cobra.Command, _ []string) error {
		if deleteFile == "" {
			return errors.New("specify a resource type (backends, frontends, server) or use -f/--file")
		}
		return deleteFromFile(deleteFile)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringVarP(&deleteFile, "file", "f", "", "Delete resource from YAML manifest (kind: Backend, Frontend, or Server)")

	// Add subcommands.
	deleteCmd.AddCommand(backends.DeleteBackendsCmd)
	deleteCmd.AddCommand(certificates.DeleteCertificatesCmd)
	deleteCmd.AddCommand(servers.DeleteServersCmd)
	deleteCmd.AddCommand(frontends.DeleteFrontendsCmd)
	deleteCmd.AddCommand(userlists.DeleteUserlistsCmd)
}

func deleteFromFile(filepath string) error {
	// The CLI is expected to read user-specified manifest files.
	data, err := os.ReadFile(filepath) //nolint:gosec // filepath comes from user input by design
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filepath, err)
	}

	var meta struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Name       string `yaml:"name"`
		Backend    string `yaml:"backend,omitempty"`
		Parent     string `yaml:"parent,omitempty"`
	}

	if err := yaml.Unmarshal(data, &meta); err != nil {
		return fmt.Errorf("failed to parse YAML file: %w", err)
	}

	if meta.APIVersion != "haproxyctl/v1" {
		return fmt.Errorf("unsupported apiVersion %q (expected haproxyctl/v1)", meta.APIVersion)
	}

	if meta.Name == "" {
		return errors.New("manifest is missing required name field")
	}

	switch strings.ToLower(meta.Kind) {
	case "backend":
		return deleteBackendByName(meta.Name)
	case "frontend":
		return deleteFrontendByName(meta.Name)
	case "userlist":
		return userlists.DeleteUserlistByName(meta.Name)
	case "server":
		backendName := meta.Parent
		if backendName == "" {
			backendName = meta.Backend
		}
		if backendName == "" {
			return errors.New("server manifest must specify backend or parent")
		}
		return servers.DeleteServer(backendName, meta.Name)
	default:
		return fmt.Errorf("unsupported resource kind: %s (supported: Backend, Frontend, Server, Userlist)", meta.Kind)
	}
}

func deleteBackendByName(name string) error {
	version, err := internal.GetConfigurationVersion()
	if err != nil {
		return fmt.Errorf("failed to fetch HAProxy configuration version: %w", err)
	}

	endpoint := "/services/haproxy/configuration/backends/" + name
	_, err = internal.SendRequest(
		"DELETE",
		endpoint,
		map[string]string{"version": strconv.Itoa(version)},
		nil,
	)
	if err != nil {
		return internal.FormatAPIError("Backend", name, "delete", err)
	}

	internal.PrintStatus("Backend", name, internal.ActionDeleted)
	return nil
}

func deleteFrontendByName(name string) error {
	version, err := internal.GetConfigurationVersion()
	if err != nil {
		return fmt.Errorf("failed to fetch HAProxy configuration version: %w", err)
	}

	endpoint := "/services/haproxy/configuration/frontends/" + name
	_, err = internal.SendRequest(
		"DELETE",
		endpoint,
		map[string]string{"version": strconv.Itoa(version)},
		nil,
	)
	if err != nil {
		return internal.FormatAPIError("Frontend", name, "delete", err)
	}

	internal.PrintStatus("Frontend", name, internal.ActionDeleted)
	return nil
}
