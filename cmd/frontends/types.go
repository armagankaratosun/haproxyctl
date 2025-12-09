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
	"errors"
	"fmt"
	"haproxyctl/internal"
	"log"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const sslEnabledValue = "enabled"

// BindConfig represents a single frontend bind (what the HAProxy Data Plane API expects).
type BindConfig struct {
	Address string `json:"address" yaml:"address"`
	Port    int    `json:"port"    yaml:"port"`
	SSL     bool   `json:"ssl,omitempty" yaml:"ssl,omitempty"`
	// Name is the underlying bind name in the Data Plane API.
	// It is not part of the manifest and is used only to drive
	// update/delete operations when reconciling binds.
	Name string `json:"-" yaml:"-"`
}

// bindPayload is the wire-format representation of a bind, using the
// v3 enum for ssl instead of a boolean.
type bindPayload struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	SSL     string `json:"ssl,omitempty"`
}

// toPayload converts a BindConfig into the structure expected by
// the HAProxy Data Plane API v3.
func (b BindConfig) toPayload() bindPayload {
	payload := bindPayload{
		Address: b.Address,
		Port:    b.Port,
	}
	if b.SSL {
		payload.SSL = sslEnabledValue
	}
	return payload
}

// frontendConfig is exactly what the HAProxy Data Plane API expects when creating/updating a frontend.
type frontendConfig struct {
	Name                 string            `json:"name" yaml:"name"`
	Mode                 string            `json:"mode,omitempty" yaml:"mode,omitempty"`
	DefaultBackend       string            `json:"default_backend,omitempty" yaml:"default_backend,omitempty"`
	ForwardFor           map[string]string `json:"forwardfor,omitempty" yaml:"forwardfor,omitempty"`
	TimeoutClient        string            `json:"timeout_client,omitempty" yaml:"timeout_client,omitempty"`
	TimeoutHTTPRequest   string            `json:"timeout_http_request,omitempty" yaml:"timeout_http_request,omitempty"`
	TimeoutHTTPKeepAlive string            `json:"timeout_http_keep_alive,omitempty" yaml:"timeout_http_keep_alive,omitempty"`
	TimeoutQueue         string            `json:"timeout_queue,omitempty" yaml:"timeout_queue,omitempty"`
	TimeoutServer        string            `json:"timeout_server,omitempty" yaml:"timeout_server,omitempty"`
}

// frontendPayload is the wire-format representation of a frontend,
// with timeout fields normalized to integer milliseconds for v3.
type frontendPayload struct {
	frontendConfig
	TimeoutClient        int `json:"timeout_client,omitempty"`
	TimeoutHTTPRequest   int `json:"timeout_http_request,omitempty"`
	TimeoutHTTPKeepAlive int `json:"timeout_http_keep_alive,omitempty"`
	TimeoutQueue         int `json:"timeout_queue,omitempty"`
	TimeoutServer        int `json:"timeout_server,omitempty"`
}

// populateFrontendConfigFromMap maps a generic API frontend object into
// the strongly-typed frontendConfig used by manifests.
func populateFrontendConfigFromMap(cfg *frontendConfig, obj map[string]interface{}) {
	if v, ok := obj["name"].(string); ok {
		cfg.Name = v
	}
	if v, ok := obj["mode"].(string); ok {
		cfg.Mode = v
	}
	if v, ok := obj["default_backend"].(string); ok {
		cfg.DefaultBackend = v
	}
	if m, ok := obj["forwardfor"].(map[string]interface{}); ok {
		cfg.ForwardFor = toStringMap(m)
	}

	if ms, ok := getIntField(obj, "timeout_client"); ok {
		cfg.TimeoutClient = internal.FormatMillisAsDuration(ms)
	}
	if ms, ok := getIntField(obj, "timeout_http_request"); ok {
		cfg.TimeoutHTTPRequest = internal.FormatMillisAsDuration(ms)
	}
	if ms, ok := getIntField(obj, "timeout_http_keep_alive"); ok {
		cfg.TimeoutHTTPKeepAlive = internal.FormatMillisAsDuration(ms)
	}
	if ms, ok := getIntField(obj, "timeout_queue"); ok {
		cfg.TimeoutQueue = internal.FormatMillisAsDuration(ms)
	}
	if ms, ok := getIntField(obj, "timeout_server"); ok {
		cfg.TimeoutServer = internal.FormatMillisAsDuration(ms)
	}
}

// frontendWithBinds is the user‑facing structure: includes metadata, core frontendConfig,
// plus zero or more BindConfig entries (from flags or YAML).
type frontendWithBinds struct {
	APIVersion     string `yaml:"apiVersion"`
	Kind           string `yaml:"kind"`
	frontendConfig `yaml:",inline"`
	Binds          []BindConfig `json:"binds,omitempty" yaml:"binds,omitempty"`
}

// LoadFromFile loads a YAML manifest into this struct.
func (f *frontendWithBinds) LoadFromFile(path string) error {
	data, err := internal.LoadYAMLFile(path)
	if err != nil {
		return fmt.Errorf("failed to read frontend file %q: %w", path, err)
	}
	return yaml.Unmarshal(data, f)
}

// LoadFromFlags pulls everything from Cobra flags.
func (f *frontendWithBinds) LoadFromFlags(cmd *cobra.Command, name string) {
	f.APIVersion = "haproxyctl/v1"
	f.Kind = "Frontend"
	f.Name = name

	f.Mode = internal.GetFlagString(cmd, "mode")
	f.DefaultBackend = internal.GetFlagString(cmd, "default-backend")
	f.ForwardFor = internal.GetFlagMap(cmd, "forwardfor")
	f.TimeoutClient = internal.GetFlagString(cmd, "timeout-client")
	f.TimeoutHTTPRequest = internal.GetFlagString(cmd, "timeout-http-request")
	f.TimeoutHTTPKeepAlive = internal.GetFlagString(cmd, "timeout-http-keep-alive")
	f.TimeoutQueue = internal.GetFlagString(cmd, "timeout-queue")
	f.TimeoutServer = internal.GetFlagString(cmd, "timeout-server")

	// Parse repeated --bind flags into a slice of BindConfig
	rawBinds := internal.GetFlagStringSlice(cmd, "bind")
	f.Binds = parseBindsFromFlags(rawBinds)
}

// ToFrontendConfig strips out the binds metadata so you can POST the core API object.
func (f *frontendWithBinds) ToFrontendConfig() frontendConfig {
	return f.frontendConfig
}

// ToPayload converts the CLI/YAML view into the frontendPayload that
// matches the Data Plane API v3 schema.
func (f *frontendWithBinds) ToPayload() frontendPayload {
	payload := frontendPayload{
		frontendConfig: f.frontendConfig,
	}

	if ms, err := internal.ParseDurationToMillis(f.TimeoutClient); err != nil {
		log.Fatalf("invalid frontend timeout_client: %v", err)
	} else if ms > 0 {
		payload.TimeoutClient = ms
	}

	if ms, err := internal.ParseDurationToMillis(f.TimeoutHTTPRequest); err != nil {
		log.Fatalf("invalid frontend timeout_http_request: %v", err)
	} else if ms > 0 {
		payload.TimeoutHTTPRequest = ms
	}

	if ms, err := internal.ParseDurationToMillis(f.TimeoutHTTPKeepAlive); err != nil {
		log.Fatalf("invalid frontend timeout_http_keep_alive: %v", err)
	} else if ms > 0 {
		payload.TimeoutHTTPKeepAlive = ms
	}

	if ms, err := internal.ParseDurationToMillis(f.TimeoutQueue); err != nil {
		log.Fatalf("invalid frontend timeout_queue: %v", err)
	} else if ms > 0 {
		payload.TimeoutQueue = ms
	}

	if ms, err := internal.ParseDurationToMillis(f.TimeoutServer); err != nil {
		log.Fatalf("invalid frontend timeout_server: %v", err)
	} else if ms > 0 {
		payload.TimeoutServer = ms
	}

	return payload
}

// Validate does minimal sanity checks before attempting creation.
func (f *frontendWithBinds) Validate() error {
	if f.Name == "" {
		return errors.New("frontend name is required")
	}
	if f.Mode != "http" && f.Mode != "tcp" {
		return fmt.Errorf("invalid mode %q (allowed: http, tcp)", f.Mode)
	}
	// Binds are optional; if provided, ensure address+port are set
	for _, b := range f.Binds {
		if b.Address == "" || b.Port == 0 {
			return fmt.Errorf("each bind must have address and port: %+v", b)
		}
	}
	return nil
}

// parseBindsFromFlags turns strings like "address=0.0.0.0,port=80,ssl=enabled"
// into a []BindConfig, converting port→int and ssl→bool.
func parseBindsFromFlags(flags []string) []BindConfig {
	var out []BindConfig
	for _, raw := range flags {
		parts := strings.Split(raw, ",")
		var b BindConfig
		for _, kv := range parts {
			pair := strings.SplitN(kv, "=", 2)
			if len(pair) != 2 {
				continue
			}
			key, val := pair[0], pair[1]
			switch key {
			case "address":
				b.Address = val
			case "port":
				if p, err := strconv.Atoi(val); err == nil {
					b.Port = p
				}
			case "ssl":
				b.SSL = (val == "true" || val == "enabled")
			}
		}
		if b.Address != "" && b.Port != 0 {
			out = append(out, b)
		}
	}
	return out
}

// toStringMap converts a map[string]interface{} to map[string]string.
func toStringMap(m map[string]interface{}) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		switch val := v.(type) {
		case string:
			out[k] = val
		default:
			out[k] = fmt.Sprintf("%v", val)
		}
	}
	return out
}

// getIntField extracts an integer field from a generic map where
// numbers are typically float64 from JSON decoding.
func getIntField(obj map[string]interface{}, key string) (int, bool) {
	v, ok := obj[key]
	if !ok || v == nil {
		return 0, false
	}

	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}
