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
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

// OutputFormatYAML is the canonical YAML output format string.
const OutputFormatYAML = "yaml"

// LoadYAMLFile reads the content of a YAML file.
func LoadYAMLFile(filepath string) ([]byte, error) {
	data, err := os.ReadFile(filepath) //nolint:gosec // CLI intentionally reads user-specified manifest paths
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filepath, err)
	}
	return data, nil
}

// FormatOutput prints structured data according to the requested output format.
func FormatOutput(data interface{}, outputFormat string) {
	// Normalize `[]map[string]interface{}` to `[]interface{}`.
	if v, ok := data.([]map[string]interface{}); ok {
		genericList := make([]interface{}, 0, len(v))
		for _, item := range v {
			genericList = append(genericList, item)
		}
		data = genericList
	}

	// Handle default format.
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

	// YAML and JSON (explicitly requested).
	switch outputFormat {
	case OutputFormatYAML:
		yamlOutput, err := yaml.Marshal(data)
		if err != nil {
			log.Fatalf("Failed to generate YAML: %v", err)
		}
		if _, err := fmt.Fprintln(os.Stdout, string(yamlOutput)); err != nil {
			log.Printf("warning: failed to write YAML output: %v", err)
		}

	case "json":
		jsonOutput, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			log.Fatalf("Failed to generate JSON: %v", err)
		}
		if _, err := fmt.Fprintln(os.Stdout, string(jsonOutput)); err != nil {
			log.Printf("warning: failed to write JSON output: %v", err)
		}

	default:
		log.Fatalf("Invalid output format: %s. Supported formats: yaml, json, table", outputFormat)
	}
}

// ConvertMapToTyped converts string map into map[string]interface{}.
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

// SortByStringField sorts a slice of map[string]interface{} in place by the
// given string field. Missing or non-string fields are treated as empty
// strings, so they appear first but deterministically.
func SortByStringField(list []map[string]interface{}, field string) {
	sort.Slice(list, func(i, j int) bool {
		nameI, _ := list[i][field].(string)
		nameJ, _ := list[j][field].(string)
		return nameI < nameJ
	})
}

// PrintResourceDescription prints structured details for a resource (backend, frontend, etc.).
func PrintResourceDescription(resourceType string, resource map[string]interface{}, sections map[string][]string, servers []map[string]interface{}) {
	if _, err := fmt.Fprintf(os.Stdout, "%s: %s\n", resourceType, resource["name"]); err != nil {
		log.Printf("warning: failed to write resource header: %v", err)
	}

	if basics, exists := sections["basic"]; exists {
		for _, field := range basics {
			if value, ok := resource[field]; ok && value != "" {
				if _, err := fmt.Fprintf(os.Stdout, "%s: %v\n", formatFieldName(field), value); err != nil {
					log.Printf("warning: failed to write basic field: %v", err)
				}
			}
		}
	}

	for sectionName, fields := range sections {
		if sectionName == "basic" {
			continue
		}
		if _, err := fmt.Fprintf(os.Stdout, "\n%s:\n", cases.Title(language.English).String(sectionName)); err != nil {
			log.Printf("warning: failed to write section header: %v", err)
		}
		for _, field := range fields {
			if value, ok := resource[field]; ok && value != "" {
				if _, err := fmt.Fprintf(os.Stdout, "- %s: %v\n", formatFieldName(field), value); err != nil {
					log.Printf("warning: failed to write section field: %v", err)
				}
			}
		}
	}

	if len(servers) > 0 {
		if _, err := fmt.Fprintln(os.Stdout, "\nServers:"); err != nil {
			log.Printf("warning: failed to write servers header: %v", err)
		}
		const (
			tabWidth   = 8
			tabPadding = 2
		)

		w := tabwriter.NewWriter(os.Stdout, 0, tabWidth, tabPadding, ' ', 0)
		if _, err := fmt.Fprintf(w, "NAME\tADDRESS\tPORT\tWEIGHT\n"); err != nil {
			log.Printf("warning: failed to write servers header: %v", err)
		}
		if _, err := fmt.Fprintf(w, "----\t-------\t----\t------\n"); err != nil {
			log.Printf("warning: failed to write servers separator: %v", err)
		}
		for _, server := range servers {
			if _, err := fmt.Fprintf(w, "%s\t%s\t%v\t%v\n",
				server["name"],
				server["address"],
				server["port"],
				server["weight"],
			); err != nil {
				log.Printf("warning: failed to write server row: %v", err)
			}
		}
		if err := w.Flush(); err != nil {
			log.Printf("warning: failed to flush servers table: %v", err)
		}
	}
}

// printTable formats structured data into a clean table like kubectl.
func printTable(data []interface{}) {
	if len(data) == 0 {
		if _, err := fmt.Fprintln(os.Stdout, "No resources found."); err != nil {
			log.Printf("warning: failed to write empty-table message: %v", err)
		}
		return
	}

	firstRow, ok := data[0].(map[string]interface{})
	if !ok {
		if _, err := fmt.Fprintln(os.Stdout, "Invalid data format."); err != nil {
			log.Printf("warning: failed to write invalid-data message: %v", err)
		}
		return
	}

	headers := getSortedKeys(firstRow)

	const (
		printTabWidth   = 8
		printTabPadding = 2
	)

	w := tabwriter.NewWriter(os.Stdout, 0, printTabWidth, printTabPadding, ' ', 0)

	for _, key := range headers {
		if _, err := fmt.Fprintf(w, "%s\t", strings.ToUpper(key)); err != nil {
			log.Printf("warning: failed to write table header: %v", err)
			return
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		log.Printf("warning: failed to terminate header line: %v", err)
		return
	}

	for range headers {
		if _, err := fmt.Fprintf(w, "--------\t"); err != nil {
			log.Printf("warning: failed to write header separator: %v", err)
			return
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		log.Printf("warning: failed to terminate separator line: %v", err)
		return
	}

	for _, row := range data {
		rowMap, ok := row.(map[string]interface{})
		if !ok {
			continue
		}
		for _, key := range headers {
			if _, err := fmt.Fprintf(w, "%v\t", formatValue(rowMap[key])); err != nil {
				log.Printf("warning: failed to write table value: %v", err)
				return
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			log.Printf("warning: failed to terminate row line: %v", err)
			return
		}
	}

	if err := w.Flush(); err != nil {
		log.Printf("warning: failed to flush table: %v", err)
	}
}

// getSortedKeys ensures "NAME" or similar fields appear first, with the rest sorted alphabetically.
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

// formatFieldName makes fields look prettier (timeout_client -> Timeout client).
func formatFieldName(name string) string {
	return cases.Title(language.English).String(strings.ReplaceAll(name, "_", " "))
}

// formatValue ensures human-readable formatting for values.
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
		return strconv.FormatBool(v) // Print true/false
	case []interface{}:
		return formatList(v) // <- Handle lists (like servers)
	case map[string]interface{}:
		return formatNestedMap(v) // existing handling
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatList formats list values for table output.
func formatList(list []interface{}) string {
	if len(list) == 0 {
		return "-"
	}

	// Special case: detect if this is a list of servers.
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

	// Fallback for generic lists (non-servers).
	return fmt.Sprintf("%v", list)
}

func isServerObject(item map[string]interface{}) bool {
	_, hasName := item["name"]
	_, hasAddress := item["address"]
	_, hasPort := item["port"]
	return hasName && hasAddress && hasPort
}

// formatNestedMap handles nested structures (like balance.algorithm).
func formatNestedMap(nested map[string]interface{}) string {
	if len(nested) == 1 {
		for _, v := range nested {
			return formatValue(v)
		}
	}
	return "{...}"
}

// ParseDurationToMillis normalizes user-friendly duration strings
// (e.g. "30s", "500ms") or plain integers (already in ms) into a
// millisecond value suitable for the Data Plane API v3 timeout fields.
func ParseDurationToMillis(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}

	// If it's a bare integer, assume it's already milliseconds.
	if n, err := strconv.Atoi(s); err == nil {
		return n, nil
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", s, err)
	}
	return int(d / time.Millisecond), nil
}

// FormatMillisAsDuration renders a millisecond value as a human‑readable
// duration string (e.g. 30000 -> "30s"). A zero value returns an empty
// string so omitted timeouts remain omitted in manifests.
func FormatMillisAsDuration(ms int) string {
	if ms == 0 {
		return ""
	}
	d := time.Duration(ms) * time.Millisecond
	return d.String()
}

// ManifestList represents a generic list of manifest-style resources,
// similar to the Kubernetes List type. It is used only for structured
// YAML/JSON output (e.g. get ... -o yaml) and is not used in table mode.
type ManifestList struct {
	APIVersion string        `json:"apiVersion" yaml:"apiVersion"`
	Kind       string        `json:"kind" yaml:"kind"`
	Items      []interface{} `json:"items" yaml:"items"`
}
