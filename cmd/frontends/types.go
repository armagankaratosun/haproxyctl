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
package frontends

import (
	"fmt"
	"strconv"
	"strings"

	"haproxyctl/internal"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// BindConfig represents a single frontend bind (what the HAProxy Data Plane API expects)
type BindConfig struct {
	Address string `json:"address" yaml:"address"`
	Port    int    `json:"port"    yaml:"port"`
	SSL     bool   `json:"ssl,omitempty" yaml:"ssl,omitempty"`
}

// frontendConfig is exactly what the HAProxy Data Plane API expects when creating/updating a frontend
type frontendConfig struct {
	Name                 string            `json:"name" yaml:"name"`
	Mode                 string            `json:"mode,omitempty" yaml:"mode,omitempty"`
	DefaultBackend       string            `json:"default_backend,omitempty" yaml:"default_backend,omitempty"`
	ForwardFor           map[string]string `json:"forwardfor,omitempty" yaml:"forwardfor,omitempty"`
	TimeoutClient        string            `json:"timeout_client,omitempty" yaml:"timeout_client,omitempty"`
	TimeoutHttpRequest   string            `json:"timeout_http_request,omitempty" yaml:"timeout_http_request,omitempty"`
	TimeoutHttpKeepAlive string            `json:"timeout_http_keep_alive,omitempty" yaml:"timeout_http_keep_alive,omitempty"`
	TimeoutQueue         string            `json:"timeout_queue,omitempty" yaml:"timeout_queue,omitempty"`
	TimeoutServer        string            `json:"timeout_server,omitempty" yaml:"timeout_server,omitempty"`
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
	f.TimeoutHttpRequest = internal.GetFlagString(cmd, "timeout-http-request")
	f.TimeoutHttpKeepAlive = internal.GetFlagString(cmd, "timeout-http-keep-alive")
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

// Validate does minimal sanity checks before attempting creation.
func (f *frontendWithBinds) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("frontend name is required")
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
