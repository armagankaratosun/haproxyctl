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
package utils

import (
	"github.com/spf13/cobra"
)

// OverrideFlag updates a map with CLI flag values, if the flag is set.
func OverrideFlag(cmd *cobra.Command, backendData map[string]interface{}, flag string, jsonKey ...string) {
	key := flag
	if len(jsonKey) > 0 {
		key = jsonKey[0] // Allows mapping flag names to JSON keys
	}

	if cmd.Flags().Changed(flag) {
		val, _ := cmd.Flags().GetString(flag)
		if val != "" {
			backendData[key] = val
		}
	}
}

// OverrideFlagInt is specifically for integer flags.
func OverrideFlagInt(cmd *cobra.Command, backendData map[string]interface{}, flag string, jsonKey ...string) {
	key := flag
	if len(jsonKey) > 0 {
		key = jsonKey[0]
	}

	if cmd.Flags().Changed(flag) {
		val, _ := cmd.Flags().GetInt(flag)
		backendData[key] = val
	}
}
