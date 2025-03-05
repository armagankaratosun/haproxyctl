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
	"errors"
	"fmt"
	"log"
	"strings"
)

// ResourceAliases maps singular and plural forms to the canonical plural form
var ResourceAliases = map[string]string{
	"backend":       "backends",
	"backends":      "backends",
	"server":        "servers",
	"servers":       "servers",
	"frontend":      "frontends",
	"frontends":     "frontends",
	"acl":           "acls",
	"acls":          "acls",
	"configuration": "configuration", // No plural for configuration
}

// NormalizeResource converts singular to plural if needed, and validates the resource
func NormalizeResource(resource string) (string, error) {
	lower := strings.ToLower(resource)
	if normalized, exists := ResourceAliases[lower]; exists {
		return normalized, nil
	}
	return "", errors.New("unknown resource type: " + resource)
}

// EnrichBackendWithServers fetches and attaches servers to a backend object as []interface{}
func EnrichBackendWithServers(backend map[string]interface{}) {
	backendName, ok := backend["name"].(string)
	if !ok || backendName == "" {
		log.Fatalf("Backend has no valid name field: %+v", backend)
	}

	endpoint := fmt.Sprintf("/services/haproxy/configuration/backends/%s/servers", backendName)
	data, err := SendRequest("GET", endpoint, nil, nil)
	if err != nil {
		log.Fatalf("Failed to fetch servers for backend %s: %v", backendName, err)
	}

	var servers []map[string]interface{}
	if err := json.Unmarshal(data, &servers); err != nil {
		log.Fatalf("Failed to parse servers response: %v\nResponse: %s", err, string(data))
	}

	var serverInterfaces []interface{}
	for _, server := range servers {
		if !isServerObject(server) {
			log.Fatalf("Unexpected server format in backend %s: %+v", backendName, server)
		}
		serverInterfaces = append(serverInterfaces, server)
	}

	// Attach as []interface{}, which plays nicely with formatList()
	backend["servers"] = serverInterfaces
}

// ValidateBackend performs basic validation for backend fields
func ValidateBackend(backend map[string]interface{}) error {
	requiredFields := []string{"name", "mode", "balance"}

	for _, field := range requiredFields {
		if value, exists := backend[field]; !exists || value == "" {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	balance, ok := backend["balance"].(map[string]interface{})
	if !ok || balance["algorithm"] == "" {
		return errors.New("missing or invalid balance.algorithm field")
	}

	// Optional deeper validation (you could extend this if needed)
	if defaultServer, ok := backend["default_server"].(map[string]interface{}); ok {
		if maxconn, exists := defaultServer["maxconn"]; exists {
			if _, ok := maxconn.(int); !ok {
				return errors.New("default_server.maxconn should be an integer")
			}
		}
	}

	return nil
}
