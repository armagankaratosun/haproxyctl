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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const httpErrorThreshold = 300

// ParseAPIResponse unmarshals raw API response bytes into the provided target.
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

// SendRequest is a generic function to send API requests.
func SendRequest(method, endpoint string, queryParams map[string]string, body interface{}) ([]byte, error) {
	return SendRequestWithContext(context.Background(), method, endpoint, queryParams, body)
}

// SendRequestWithContext sends an API request using the provided context.
// Most callers should prefer this so that requests can be cancelled when
// the associated CLI command is cancelled.
func SendRequestWithContext(ctx context.Context, method, endpoint string, queryParams map[string]string, body interface{}) ([]byte, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	// Build URL with query parameters
	baseURL := normalizeAPIBaseURL(cfg.APIBaseURL)
	url := baseURL + endpoint
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
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create API request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(cfg.Username, cfg.Password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close response body: %v\n", cerr)
		}
	}()

	if resp.StatusCode >= httpErrorThreshold {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("HAProxy API error (%d) and failed to read error body: %w", resp.StatusCode, readErr)
		}
		return nil, fmt.Errorf("HAProxy API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	return io.ReadAll(resp.Body)
}

// SendRawRequest sends a raw payload (e.g. entire HAProxy config) without JSON‑encoding.
// contentType should be "text/plain" or "application/octet-stream".
func SendRawRequest(method, endpoint string, queryParams map[string]string, rawBody []byte, contentType string) ([]byte, error) {
	return SendRawRequestWithContext(context.Background(), method, endpoint, queryParams, rawBody, contentType)
}

// SendRawRequestWithContext is the context-aware form of SendRawRequest.
func SendRawRequestWithContext(ctx context.Context, method, endpoint string, queryParams map[string]string, rawBody []byte, contentType string) ([]byte, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	// Build URL + query string
	baseURL := normalizeAPIBaseURL(cfg.APIBaseURL)
	url := baseURL + endpoint
	if len(queryParams) > 0 {
		q := "?"
		for k, v := range queryParams {
			q += fmt.Sprintf("%s=%s&", k, v)
		}
		url += strings.TrimSuffix(q, "&")
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(rawBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.SetBasicAuth(cfg.Username, cfg.Password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("raw request failed: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close response body: %v\n", cerr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read raw response body: %w", err)
	}
	if resp.StatusCode >= httpErrorThreshold {
		return nil, fmt.Errorf("HAProxy API error (%d): %s", resp.StatusCode, string(body))
	}
	return body, nil
}

// normalizeAPIBaseURL ensures the configured API base URL includes a version
// prefix (we default to v3) and has no trailing slash. This lets users enter
// either "http://host:5555" or "http://host:5555/v3" in `haproxyctl login`.
func normalizeAPIBaseURL(raw string) string {
	base := strings.TrimSpace(raw)
	base = strings.TrimRight(base, "/")

	if strings.HasSuffix(base, "/v1") ||
		strings.HasSuffix(base, "/v2") ||
		strings.HasSuffix(base, "/v3") {
		return base
	}

	// Default to Data Plane API v3 when no explicit version is present.
	return base + "/v3"
}

// IsNotFoundError reports whether the given error corresponds to a 404
// response from the Data Plane API. It inspects the wrapped error message
// produced by SendRequest.
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	// SendRequest formats errors as: "HAProxy API error (%d): ..."
	return strings.Contains(err.Error(), "HAProxy API error (404)")
}

// GetResource retrieves a single resource (map[string]interface{}) from the API.
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

// GetResourceList retrieves a list of resources ([]map[string]interface{}) from the API.
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

// UploadSSLCertificateWithContext uploads a PEM bundle (key + cert + optional
// chain) to the HAProxy Data Plane API ssl_certificates storage.
func UploadSSLCertificateWithContext(ctx context.Context, name string, pem []byte) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	baseURL := normalizeAPIBaseURL(cfg.APIBaseURL)
	url := baseURL + "/services/haproxy/storage/ssl_certificates"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	filename := name + ".pem"
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return fmt.Errorf("failed to create multipart form file: %w", err)
	}

	if _, err := part.Write(pem); err != nil {
		return fmt.Errorf("failed to write PEM data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to finalize multipart body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return fmt.Errorf("failed to create SSL certificate upload request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetBasicAuth(cfg.Username, cfg.Password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("SSL certificate upload failed: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close response body: %v\n", cerr)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read SSL certificate upload response: %w", err)
	}

	if resp.StatusCode >= httpErrorThreshold {
		return fmt.Errorf("HAProxy API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// UploadSSLCertificate is a convenience wrapper around UploadSSLCertificateWithContext
// using a background context.
func UploadSSLCertificate(name string, pem []byte) error {
	return UploadSSLCertificateWithContext(context.Background(), name, pem)
}
