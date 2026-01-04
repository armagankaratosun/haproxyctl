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

// Package frontends provides commands to manage HAProxy frontends.
package frontends

import (
	"fmt"
	"haproxyctl/internal"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// DescribeFrontendsCmd represents "describe frontends".
var DescribeFrontendsCmd = &cobra.Command{
	Use:   "frontends <frontend_name>",
	Short: "Describe a specific HAProxy frontend and its binds",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		frontendName := args[0]
		describeFrontend(frontendName)
	},
}

// describeFrontend fetches a frontend and its binds, and prints a detailed description.
func describeFrontend(frontendName string) {
	frontend, err := internal.GetResource("/services/haproxy/configuration/frontends/" + frontendName)
	if err != nil {
		log.Fatalf("Failed to fetch frontend '%s': %v", frontendName, err)
	}

	// Attach binds to the frontend object so they can be shown in the "listeners" section.
	internal.EnrichFrontendWithBinds(frontend)

	httpRequestRules := fetchFrontendListSection(frontendName, "HTTP request rules", "/services/haproxy/configuration/frontends/"+frontendName+"/http_request_rules")
	httpResponseRules := fetchFrontendListSection(frontendName, "HTTP response rules", "/services/haproxy/configuration/frontends/"+frontendName+"/http_response_rules")
	tcpRequestRules := fetchFrontendListSection(frontendName, "TCP request rules", "/services/haproxy/configuration/frontends/"+frontendName+"/tcp_request_rules")

	internal.PrintResourceDescription("Frontend", frontend, frontendDescriptionSections(), nil)

	printRuleSection("HTTP Request Rules", httpRequestRules)
	printRuleSection("HTTP Response Rules", httpResponseRules)
	printRuleSection("TCP Request Rules", tcpRequestRules)
}

// frontendDescriptionSections defines the sections and fields to display in frontend descriptions.
func frontendDescriptionSections() map[string][]string {
	return map[string][]string{
		"basic":     {"name", "mode", "default_backend"},
		"listeners": {"binds"},
		"timeouts":  {"timeout_client", "timeout_http_request", "timeout_http_keep_alive"},
		"logging":   {"log"},
		"options":   {"forwardfor"},
	}
}

// fetchFrontendListSection retrieves a list-valued configuration section for a frontend,
// such as HTTP/TCP rules. Failures are logged as warnings so that describe output
// remains as complete as possible.
func fetchFrontendListSection(frontendName, sectionLabel, endpoint string) []map[string]interface{} {
	list, err := internal.GetResourceList(endpoint)
	if err != nil {
		if !internal.IsNotFoundError(err) {
			log.Printf("warning: failed to fetch %s for frontend %q: %v", sectionLabel, frontendName, err)
		}
		return nil
	}
	return list
}

// printRuleSection renders a list of rule/check objects as a simple, readable section.
func printRuleSection(title string, rules []map[string]interface{}) {
	if len(rules) == 0 {
		return
	}

	if _, err := fmt.Fprintf(os.Stdout, "\n%s:\n", title); err != nil {
		log.Printf("warning: failed to write %s header: %v", title, err)
		return
	}

	for _, rule := range rules {
		line := summarizeRule(rule)
		if line == "" {
			continue
		}
		if _, err := fmt.Fprintf(os.Stdout, "- %s\n", line); err != nil {
			log.Printf("warning: failed to write %s line: %v", title, err)
			return
		}
	}
}

// summarizeRule builds a compact one-line summary for a rule/check object.
func summarizeRule(rule map[string]interface{}) string {
	var parts []string

	if idx, ok := extractInt(rule, "index"); ok {
		parts = append(parts, fmt.Sprintf("#%d", idx))
	}
	if typ := extractString(rule, "type"); typ != "" {
		parts = append(parts, typ)
	}

	cond := extractString(rule, "cond")
	condTest := extractString(rule, "cond_test")
	if cond != "" {
		if condTest != "" {
			parts = append(parts, fmt.Sprintf("%s %s", cond, condTest))
		} else {
			parts = append(parts, cond)
		}
	}

	if uri := extractString(rule, "uri"); uri != "" {
		parts = append(parts, "uri="+uri)
	}
	if port, ok := extractInt(rule, "port"); ok {
		parts = append(parts, fmt.Sprintf("port=%d", port))
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " | ")
}

func extractString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func extractInt(m map[string]interface{}, key string) (int, bool) {
	v, ok := m[key]
	if !ok {
		return 0, false
	}

	switch t := v.(type) {
	case float64:
		return int(t), true
	case int:
		return t, true
	case int64:
		return int(t), true
	case string:
		n, err := strconv.Atoi(t)
		if err == nil {
			return n, true
		}
	}
	return 0, false
}
