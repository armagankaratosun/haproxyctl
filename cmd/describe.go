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
	"log"

	"haproxyctl/cmd/backends"
	"haproxyctl/cmd/frontends"
	"haproxyctl/cmd/servers"

	"github.com/spf13/cobra"
)

// describeCmd represents the describe command.
var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe resources in HAProxy",
	Run: func(_ *cobra.Command, _ []string) {
		log.Fatal("Specify a resource type (backends, servers).")
	},
}

func init() {
	rootCmd.AddCommand(describeCmd)

	// Add subcommands.
	describeCmd.AddCommand(backends.DescribeBackendsCmd)
	describeCmd.AddCommand(frontends.DescribeFrontendsCmd)
	describeCmd.AddCommand(servers.DescribeServersCmd)
}
