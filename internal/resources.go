// Package internal contains shared helpers for haproxyctl.
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
	"errors"
	"fmt"
	"log"
	"net/url"
)

// ExtractOptionalArg extracts the optional second argument (resource name), if provided.
func ExtractOptionalArg(args []string) string {
	if len(args) > 1 {
		return args[1]
	}
	return ""
}

// EnrichBackendWithServers fetches and attaches servers to a backend object as []interface{}.
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

	serverInterfaces := make([]interface{}, 0, len(servers))
	for _, server := range servers {
		if !isServerObject(server) {
			log.Fatalf("Unexpected server format in backend %s: %+v", backendName, server)
		}
		serverInterfaces = append(serverInterfaces, server)
	}

	// Attach as []interface{}, which plays nicely with formatList().
	backend["servers"] = serverInterfaces
}

// EnrichFrontendWithBinds fetches and attaches binds to a frontend object as []interface{}.
// This allows table output to render a compact binds column similar to how servers
// are shown for backends.
func EnrichFrontendWithBinds(frontend map[string]interface{}) {
	frontendName, ok := frontend["name"].(string)
	if !ok || frontendName == "" {
		log.Fatalf("Frontend has no valid name field: %+v", frontend)
	}

	// Encode the frontend name in case it contains characters that need escaping.
	escapedName := url.PathEscape(frontendName)

	endpoint := fmt.Sprintf("/services/haproxy/configuration/frontends/%s/binds", escapedName)
	data, err := SendRequest("GET", endpoint, nil, nil)
	if err != nil {
		log.Fatalf("Failed to fetch binds for frontend %s: %v", frontendName, err)
	}

	var binds []map[string]interface{}
	if err := json.Unmarshal(data, &binds); err != nil {
		log.Fatalf("Failed to parse binds response: %v\nResponse: %s", err, string(data))
	}

	bindInterfaces := make([]interface{}, 0, len(binds))
	for _, bind := range binds {
		bindInterfaces = append(bindInterfaces, bind)
	}

	frontend["binds"] = bindInterfaces
}

// ValidateBackend performs basic validation for backend fields.
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
