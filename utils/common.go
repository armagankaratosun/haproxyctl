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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v2"
)

// FormatOutput formats data as YAML, JSON, or a table if no format is specified
func FormatOutput(data interface{}, outputFormat string) {
	// If the data is a JSON string, parse it before formatting
	var parsedData interface{}

	switch v := data.(type) {
	case string:
		// Try to unmarshal the JSON string into a structured object
		if err := json.Unmarshal([]byte(v), &parsedData); err != nil {
			parsedData = v // If parsing fails, just print raw string
		}
	default:
		parsedData = v
	}

	// Detect if parsedData is a single object `{}` and convert it to a list `[]` for table formatting
	if outputFormat == "" {
		switch v := parsedData.(type) {
		case map[string]interface{}:
			printTable([]interface{}{v}) // Convert to list and print as a table
			return
		case []interface{}:
			printTable(v) // Print list as a table
			return
		}
		outputFormat = "json" // Default to JSON for non-table cases
	}

	// Handle JSON/YAML output
	switch outputFormat {
	case "yaml":
		yamlOutput, err := yaml.Marshal(parsedData)
		if err != nil {
			log.Fatal("Failed to generate YAML:", err)
		}
		fmt.Println(string(yamlOutput))
	case "json":
		jsonOutput, err := json.MarshalIndent(parsedData, "", "    ")
		if err != nil {
			log.Fatal("Failed to generate JSON:", err)
		}
		fmt.Println(string(jsonOutput))
	default:
		log.Fatalf("Invalid output format: %s. Supported formats: yaml, json", outputFormat)
	}
}

// printTable formats structured data into a clean table like kubectl
func printTable(data []interface{}) {
	if len(data) == 0 {
		fmt.Println("No resources found.")
		return
	}

	// Extract column headers from the first object
	firstRow, ok := data[0].(map[string]interface{})
	if !ok {
		fmt.Println("Invalid data format.")
		return
	}

	// Determine and sort headers
	headers := getSortedKeys(firstRow)

	// Create a tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

	// Print column headers in uppercase like kubectl
	for _, key := range headers {
		fmt.Fprintf(w, "%s\t", strings.ToUpper(key))
	}
	fmt.Fprintln(w)

	// Print a separator line
	for range headers {
		fmt.Fprintf(w, "--------\t")
	}
	fmt.Fprintln(w)

	// Print rows with aligned values
	for _, row := range data {
		rowMap, ok := row.(map[string]interface{})
		if !ok {
			continue
		}
		for _, key := range headers {
			fmt.Fprintf(w, "%v\t", formatValue(rowMap[key]))
		}
		fmt.Fprintln(w)
	}

	// Flush output
	w.Flush()
}

// getSortedKeys ensures "NAME" or similar fields appear first, with the rest sorted alphabetically
func getSortedKeys(row map[string]interface{}) []string {
	priorityFields := []string{"name", "acl_name", "id"} // Prioritize these fields

	var primaryColumn string
	var otherColumns []string

	for key := range row {
		lowerKey := strings.ToLower(key)
		isPriority := false

		for _, priority := range priorityFields {
			if lowerKey == priority {
				primaryColumn = key
				isPriority = true
				break
			}
		}

		if !isPriority {
			otherColumns = append(otherColumns, key)
		}
	}

	sort.Strings(otherColumns) // Sort remaining fields alphabetically

	// Ensure the detected "name-like" column is first
	if primaryColumn != "" {
		return append([]string{primaryColumn}, otherColumns...)
	}

	// If no priority field is found, return sorted fields normally
	return otherColumns
}

// formatValue ensures human-readable formatting for values
func formatValue(value interface{}) string {
	if value == nil {
		return "-"
	}
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v) // Remove decimals if number
	case bool:
		return fmt.Sprintf("%t", v) // Print true/false
	case map[string]interface{}:
		return formatNestedMap(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatNestedMap handles nested structures (like balance.algorithm)
func formatNestedMap(nested map[string]interface{}) string {
	// Try to extract useful info (like `balance.algorithm`)
	if len(nested) == 1 {
		for _, v := range nested {
			return formatValue(v)
		}
	}
	return "{...}" // If more than one, just show `{...}`
}
