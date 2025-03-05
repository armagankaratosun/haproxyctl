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
	"fmt"
	"log"
	"strings"

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

// ParseKeyValueString parses "key=value,key=value" formatted strings into a map.
func ParseKeyValueString(input string) (map[string]string, error) {
	result := make(map[string]string)
	pairs := strings.Split(input, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid key=value pair: %s", pair)
		}
		result[parts[0]] = parts[1]
	}
	return result, nil
}

// ParseKeyValueFlag parses a Cobra flag containing key=value pairs
func ParseKeyValueFlag(cmd *cobra.Command, flagName string) map[string]string {
	values, err := cmd.Flags().GetStringToString(flagName)
	if err != nil {
		log.Fatalf("Failed to parse flag %s: %v", flagName, err)
	}
	if len(values) == 0 {
		return nil // Return nil if no values provided
	}
	return values
}

// GetFlagString fetches a string flag value
func GetFlagString(cmd *cobra.Command, name string) string {
	value, err := cmd.Flags().GetString(name)
	if err != nil {
		log.Fatalf("Failed to read flag %s: %v", name, err)
	}
	return value
}

// GetFlagStringSlice fetches a string slice flag value (for repeated flags like --servers)
func GetFlagStringSlice(cmd *cobra.Command, name string) []string {
	values, err := cmd.Flags().GetStringSlice(name)
	if err != nil {
		log.Fatalf("Failed to read flag %s: %v", name, err)
	}
	return values
}

// GetFlagBool fetches a boolean flag value
func GetFlagBool(cmd *cobra.Command, name string) bool {
	value, err := cmd.Flags().GetBool(name)
	if err != nil {
		log.Fatalf("Failed to read flag %s: %v", name, err)
	}
	return value
}

// GetFlagMap fetches a map flag value (key=value pairs)
func GetFlagMap(cmd *cobra.Command, name string) map[string]string {
	values, err := cmd.Flags().GetStringToString(name)
	if err != nil {
		log.Fatalf("Failed to read flag %s: %v", name, err)
	}
	return values
}

// GetFlagMapInterface fetches a map flag value and converts values to interface{} for compatibility
func GetFlagMapInterface(cmd *cobra.Command, name string) map[string]interface{} {
	values, err := cmd.Flags().GetStringToString(name)
	if err != nil {
		log.Fatalf("Failed to read flag %s: %v", name, err)
	}

	result := make(map[string]interface{})
	for k, v := range values {
		result[k] = v
	}
	return result
}

// GetFlagInt fetches an integer flag value
func GetFlagInt(cmd *cobra.Command, name string) int {
	value, err := cmd.Flags().GetInt(name)
	if err != nil {
		log.Fatalf("Failed to read flag %s: %v", name, err)
	}
	return value
}
