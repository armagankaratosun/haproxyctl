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
package backends

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"haproxyctl/cmd/servers"
	"haproxyctl/internal"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// backendConfig represents the full backend object in HAProxy Data Plane API
type backendConfig struct {
	Name                 string                   `json:"name" yaml:"name"`
	Mode                 string                   `json:"mode,omitempty" yaml:"mode,omitempty"`
	Balance              map[string]string        `json:"balance,omitempty" yaml:"balance,omitempty"`
	HTTPChkParams        map[string]string        `json:"httpchk_params,omitempty" yaml:"httpchk_params,omitempty"`
	Cookie               map[string]string        `json:"cookie,omitempty" yaml:"cookie,omitempty"`
	Log                  []map[string]interface{} `json:"log,omitempty" yaml:"log,omitempty"`
	DefaultServer        map[string]interface{}   `json:"default_server,omitempty" yaml:"default_server,omitempty"`
	ForwardFor           map[string]string        `json:"forwardfor,omitempty" yaml:"forwardfor,omitempty"`
	HTTPRequestRules     []map[string]interface{} `json:"http_request_rules,omitempty" yaml:"http_request_rules,omitempty"`
	HTTPResponseRules    []map[string]interface{} `json:"http_response_rules,omitempty" yaml:"http_response_rules,omitempty"`
	TCPRequestRules      []map[string]interface{} `json:"tcp_request_rules,omitempty" yaml:"tcp_request_rules,omitempty"`
	ErrorFiles           []map[string]interface{} `json:"error_files,omitempty" yaml:"error_files,omitempty"`
	TimeoutClient        string                   `json:"timeout_client,omitempty" yaml:"timeout_client,omitempty"`
	TimeoutHttpKeepAlive string                   `json:"timeout_http_keep_alive,omitempty" yaml:"timeout_http_keep_alive,omitempty"`
	TimeoutHttpRequest   string                   `json:"timeout_http_request,omitempty" yaml:"timeout_http_request,omitempty"`
	TimeoutQueue         string                   `json:"timeout_queue,omitempty" yaml:"timeout_queue,omitempty"`
	TimeoutServer        string                   `json:"timeout_server,omitempty" yaml:"timeout_server,omitempty"`
	TimeoutServerFin     string                   `json:"timeout_server_fin,omitempty" yaml:"timeout_server_fin,omitempty"`
	// TCPKA and Redispatch are exposed as simple booleans in the CLI/YAML
	// view, but the v3 Data Plane API expects different wire formats:
	//   - tcpka: enum "enabled"/"disabled"
	//   - redispatch: object { enabled: "enabled"/"disabled", interval: int }
	// They are therefore excluded from JSON and translated in a dedicated
	// payload struct before sending to the API.
	TCPKA      bool              `yaml:"tcpka,omitempty" json:"-"`
	Redispatch bool              `yaml:"redispatch,omitempty" json:"-"`
	Source               map[string]string        `json:"source,omitempty" yaml:"source,omitempty"`
}

// redispatchPayload matches the HAProxy Data Plane API v3 definition
// of the "redispatch" object.
type redispatchPayload struct {
	Enabled  string `json:"enabled"`
	Interval int    `json:"interval,omitempty"`
}

// backendPayload is the wire-format representation of a backend,
// embedding the base config while mapping tcpka/redispatch to their
// v3 enum/object shapes.
type backendPayload struct {
	backendConfig
	TimeoutClient        int                `json:"timeout_client,omitempty"`
	TimeoutHttpKeepAlive int                `json:"timeout_http_keep_alive,omitempty"`
	TimeoutHttpRequest   int                `json:"timeout_http_request,omitempty"`
	TimeoutQueue         int                `json:"timeout_queue,omitempty"`
	TimeoutServer        int                `json:"timeout_server,omitempty"`
	TimeoutServerFin     int                `json:"timeout_server_fin,omitempty"`
	TCPKA                string             `json:"tcpka,omitempty"`
	Redispatch           *redispatchPayload `json:"redispatch,omitempty"`
}

// backendWithServers represents the user-facing object that includes servers
// This is the structure used when reading from files or CLI flags.
type backendWithServers struct {
	APIVersion    string                 `yaml:"apiVersion"`
	Kind          string                 `yaml:"kind"`
	backendConfig `yaml:",inline"`       // Embed all backendConfig fields directly
	Servers       []servers.ServerConfig `json:"servers,omitempty" yaml:"servers,omitempty"`
}

// LoadFromFile loads backend + servers from a YAML file.
func (b *backendWithServers) LoadFromFile(filepath string) error {
	data, err := internal.LoadYAMLFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to load backend configuration file: %w", err)
	}
	if err := yaml.Unmarshal(data, b); err != nil {
		return fmt.Errorf("failed to parse backend configuration YAML: %w", err)
	}

	// Basic sanity check
	if b.Kind != "Backend" {
		return fmt.Errorf("invalid kind '%s', expected 'Backend'", b.Kind)
	}
	if b.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}

	return nil
}

// LoadFromFlags populates backend + servers from CLI flags.
func (b *backendWithServers) LoadFromFlags(cmd *cobra.Command, backendName string) {
	b.APIVersion = "haproxyctl/v1"
	b.Kind = "Backend"
	b.Name = backendName
	b.Mode = internal.GetFlagString(cmd, "mode")
	b.Balance = internal.GetFlagMap(cmd, "balance")
	b.DefaultServer = internal.GetFlagMapInterface(cmd, "default-server")
	b.ForwardFor = internal.GetFlagMap(cmd, "forwardfor")
	b.TimeoutClient = internal.GetFlagString(cmd, "timeout-client")
	b.TimeoutQueue = internal.GetFlagString(cmd, "timeout-queue")
	b.TimeoutServer = internal.GetFlagString(cmd, "timeout-server")
	b.Redispatch = internal.GetFlagBool(cmd, "redispatch")

	rawServers := internal.GetFlagStringSlice(cmd, "server")
	b.Servers = parseServersFromFlags(rawServers)
}

// parseServersFromFlags converts `--server` flags into servers.ServerConfig structs
// Example: --server name=s1,address=10.0.0.1,port=80,weight=100
func parseServersFromFlags(rawServers []string) []servers.ServerConfig {
	var result []servers.ServerConfig
	for _, serverStr := range rawServers {
		parts := strings.Split(serverStr, ",")
		server := servers.ServerConfig{}
		for _, part := range parts {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) != 2 {
				continue
			}
			key, value := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
			switch key {
			case "name":
				server.Name = value
			case "address":
				server.Address = value
			case "port":
				port, _ := strconv.Atoi(value)
				server.Port = port
			case "weight":
				weight, _ := strconv.Atoi(value)
				server.Weight = weight
			case "ssl":
				server.SSL = (strings.ToLower(value) == "true")
			}
		}
		if server.Name != "" && server.Address != "" && server.Port != 0 {
			result = append(result, server)
		}
	}
	return result
}

// ToBackendConfig strips servers (API compatibility)
func (b *backendWithServers) ToBackendConfig() backendConfig {
	return b.backendConfig
}

// toPayload converts the CLI/YAML view into the backendPayload that
// matches the Data Plane API v3 schema.
func (b *backendWithServers) toPayload() backendPayload {
	payload := backendPayload{
		backendConfig: b.backendConfig,
	}

	if ms, err := internal.ParseDurationToMillis(b.TimeoutClient); err != nil {
		log.Fatalf("invalid backend timeout_client: %v", err)
	} else if ms > 0 {
		payload.TimeoutClient = ms
	}

	if ms, err := internal.ParseDurationToMillis(b.TimeoutHttpKeepAlive); err != nil {
		log.Fatalf("invalid backend timeout_http_keep_alive: %v", err)
	} else if ms > 0 {
		payload.TimeoutHttpKeepAlive = ms
	}

	if ms, err := internal.ParseDurationToMillis(b.TimeoutHttpRequest); err != nil {
		log.Fatalf("invalid backend timeout_http_request: %v", err)
	} else if ms > 0 {
		payload.TimeoutHttpRequest = ms
	}

	if ms, err := internal.ParseDurationToMillis(b.TimeoutQueue); err != nil {
		log.Fatalf("invalid backend timeout_queue: %v", err)
	} else if ms > 0 {
		payload.TimeoutQueue = ms
	}

	if ms, err := internal.ParseDurationToMillis(b.TimeoutServer); err != nil {
		log.Fatalf("invalid backend timeout_server: %v", err)
	} else if ms > 0 {
		payload.TimeoutServer = ms
	}

	if ms, err := internal.ParseDurationToMillis(b.TimeoutServerFin); err != nil {
		log.Fatalf("invalid backend timeout_server_fin: %v", err)
	} else if ms > 0 {
		payload.TimeoutServerFin = ms
	}

	if b.TCPKA {
		payload.TCPKA = "enabled"
	}

	if b.Redispatch {
		payload.Redispatch = &redispatchPayload{
			Enabled: "enabled",
			// Interval is optional; we leave it zero-value unless
			// we add a flag to configure it.
		}
	}

	return payload
}

// Validate does basic validation for backendWithServers.
func (b *backendWithServers) Validate() error {
	if b.Name == "" {
		return fmt.Errorf("backend name is required")
	}
	if b.Kind != "Backend" {
		return fmt.Errorf("kind must be 'Backend'")
	}
	if b.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}
	if b.Mode != "http" && b.Mode != "tcp" {
		return fmt.Errorf("invalid mode: %s (allowed: http, tcp)", b.Mode)
	}
	for _, server := range b.Servers {
		if server.Name == "" || server.Address == "" || server.Port == 0 {
			return fmt.Errorf("each server must have name, address, and port")
		}
	}
	return nil
}
