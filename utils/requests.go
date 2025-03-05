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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// TODO
func ParseAPIResponse(data []byte, target interface{}) {
	err := json.Unmarshal(data, target)
	if err != nil {
		log.Fatalf("Failed to parse API response: %v\nResponse: %s", err, string(data))
	}
}

// GetConfigurationVersion retrieves the current HAProxy configuration version from the Data Plane API.
func GetConfigurationVersion() (int, error) {
	data, err := SendRequest("GET", "/services/haproxy/configuration/version", nil, nil)
	if err != nil {
		return 0, err
	}

	versionStr := strings.TrimSpace(string(data))
	versionInt, err := strconv.Atoi(versionStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse version as integer: %w", err)
	}

	return versionInt, nil
}

// SendRequest is a generic function to send API requests
func SendRequest(method, endpoint string, queryParams map[string]string, body interface{}) ([]byte, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	// Build URL with query parameters
	url := cfg.APIBase + endpoint
	if len(queryParams) > 0 {
		url += "?"
		for key, value := range queryParams {
			url += key + "=" + value + "&"
		}
		url = url[:len(url)-1] // Remove trailing "&"
	}

	// Convert body to JSON if needed
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create API request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(cfg.User, cfg.Pass)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("HAProxy API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	return ioutil.ReadAll(resp.Body)
}

// GetResource retrieves a single resource (map[string]interface{}) from the API
func GetResource(endpoint string) (map[string]interface{}, error) {
	data, err := SendRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}

	var resource map[string]interface{}
	if err := json.Unmarshal(data, &resource); err != nil {
		return nil, fmt.Errorf("failed to parse resource response: %w", err)
	}

	return resource, nil
}

// GetResourceList retrieves a list of resources ([]map[string]interface{}) from the API
func GetResourceList(endpoint string) ([]map[string]interface{}, error) {
	data, err := SendRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource list: %w", err)
	}

	var resourceList []map[string]interface{}
	if err := json.Unmarshal(data, &resourceList); err != nil {
		return nil, fmt.Errorf("failed to parse resource list response: %w", err)
	}

	return resourceList, nil
}
