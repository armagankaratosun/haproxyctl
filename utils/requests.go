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
	"net/http"
)

// sendRequest is a generic function to send API requests
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
