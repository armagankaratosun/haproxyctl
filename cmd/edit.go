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
	"haproxyctl/cmd/backends"
	"haproxyctl/cmd/configuration"
	"haproxyctl/cmd/frontends"

	"github.com/spf13/cobra"
)

// editCmd is the entrypoint for interactive editing of resources.
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit HAProxy resources in your editor",
	RunE: func(cmd *cobra.Command, _ []string) error {
		// If no subcommand is given, show help.
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.AddCommand(backends.EditBackendsCmd)
	editCmd.AddCommand(frontends.EditFrontendsCmd)
	editCmd.AddCommand(configuration.EditConfigurationCmd)
}
