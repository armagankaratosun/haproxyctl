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

package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

// LoadYAMLFile reads the content of a YAML file
func LoadYAMLFile(filepath string) ([]byte, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filepath, err)
	}
	return data, nil
}

func FormatOutput(data interface{}, outputFormat string) {
	// Normalize `[]map[string]interface{}` to `[]interface{}`
	switch v := data.(type) {
	case []map[string]interface{}:
		var genericList []interface{}
		for _, item := range v {
			genericList = append(genericList, item)
		}
		data = genericList
	}

	// Handle default format
	if outputFormat == "table" || outputFormat == "" {
		switch v := data.(type) {
		case map[string]interface{}:
			printTable([]interface{}{v}) // single object as table
			return
		case []interface{}:
			printTable(v) // list of objects as table
			return
		default:
			log.Fatalf("Cannot print table for this data type: %T", v)
		}
	}

	// YAML and JSON (explicitly requested)
	switch outputFormat {
	case "yaml":
		yamlOutput, err := yaml.Marshal(data)
		if err != nil {
			log.Fatalf("Failed to generate YAML: %v", err)
		}
		fmt.Println(string(yamlOutput))

	case "json":
		jsonOutput, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			log.Fatalf("Failed to generate JSON: %v", err)
		}
		fmt.Println(string(jsonOutput))

	default:
		log.Fatalf("Invalid output format: %s. Supported formats: yaml, json, table", outputFormat)
	}
}

// ConvertMapToTyped converts string map into map[string]interface{}
func ConvertMapToTyped(input map[string]string, intFields []string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range input {
		if Contains(intFields, k) {
			intVal, err := strconv.Atoi(v)
			if err != nil {
				log.Fatalf("Failed to convert %s to integer: %v", k, err)
			}
			result[k] = intVal
		} else {
			result[k] = v
		}
	}
	return result
}

// PrintResourceDescription prints structured details for a resource (backend, frontend, etc.)
func PrintResourceDescription(resourceType string, resource map[string]interface{}, sections map[string][]string, servers []map[string]interface{}) {
	fmt.Printf("%s: %s\n", resourceType, resource["name"])

	if basics, exists := sections["basic"]; exists {
		for _, field := range basics {
			if value, ok := resource[field]; ok && value != "" {
				fmt.Printf("%s: %v\n", formatFieldName(field), value)
			}
		}
	}

	for sectionName, fields := range sections {
		if sectionName == "basic" {
			continue
		}
		fmt.Printf("\n%s:\n", cases.Title(language.English).String(sectionName))
		for _, field := range fields {
			if value, ok := resource[field]; ok && value != "" {
				fmt.Printf("- %s: %v\n", formatFieldName(field), value)
			}
		}
	}

	if len(servers) > 0 {
		fmt.Println("\nServers:")
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
		fmt.Fprintf(w, "NAME\tADDRESS\tPORT\tWEIGHT\n")
		fmt.Fprintf(w, "----\t-------\t----\t------\n")
		for _, server := range servers {
			fmt.Fprintf(w, "%s\t%s\t%v\t%v\n",
				server["name"],
				server["address"],
				server["port"],
				server["weight"],
			)
		}
		w.Flush()
	}
}

// printTable formats structured data into a clean table like kubectl
func printTable(data []interface{}) {
	if len(data) == 0 {
		fmt.Println("No resources found.")
		return
	}

	firstRow, ok := data[0].(map[string]interface{})
	if !ok {
		fmt.Println("Invalid data format.")
		return
	}

	headers := getSortedKeys(firstRow)

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

	for _, key := range headers {
		fmt.Fprintf(w, "%s\t", strings.ToUpper(key))
	}
	fmt.Fprintln(w)

	for range headers {
		fmt.Fprintf(w, "--------\t")
	}
	fmt.Fprintln(w)

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

	w.Flush()
}

// getSortedKeys ensures "NAME" or similar fields appear first, with the rest sorted alphabetically
func getSortedKeys(row map[string]interface{}) []string {
	priorityFields := []string{"name", "acl_name", "id"}

	var primaryColumn string
	var otherColumns []string

	for key := range row {
		lowerKey := strings.ToLower(key)
		if Contains(priorityFields, lowerKey) {
			primaryColumn = key
		} else {
			otherColumns = append(otherColumns, key)
		}
	}

	sort.Strings(otherColumns)

	if primaryColumn != "" {
		return append([]string{primaryColumn}, otherColumns...)
	}

	return otherColumns
}

// formatFieldName makes fields look prettier (timeout_client -> Timeout client)
func formatFieldName(name string) string {
	return cases.Title(language.English).String(strings.ReplaceAll(name, "_", " "))
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
	case []interface{}:
		return formatList(v) // <- Handle lists (like servers)
	case map[string]interface{}:
		return formatNestedMap(v) // existing handling
	default:
		return fmt.Sprintf("%v", v)
	}
}

// TODO
func formatList(list []interface{}) string {
	if len(list) == 0 {
		return "-"
	}

	// Special case: detect if this is a list of servers
	if first, ok := list[0].(map[string]interface{}); ok && isServerObject(first) {
		var servers []string
		for _, item := range list {
			server, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			servers = append(servers, fmt.Sprintf("%s (%s:%v)", server["name"], server["address"], server["port"]))
		}
		return strings.Join(servers, ", ")
	}

	// Fallback for generic lists (non-servers)
	return fmt.Sprintf("%v", list)
}

func isServerObject(item map[string]interface{}) bool {
	_, hasName := item["name"]
	_, hasAddress := item["address"]
	_, hasPort := item["port"]
	return hasName && hasAddress && hasPort
}

// formatNestedMap handles nested structures (like balance.algorithm)
func formatNestedMap(nested map[string]interface{}) string {
	if len(nested) == 1 {
		for _, v := range nested {
			return formatValue(v)
		}
	}
	return "{...}"
}
