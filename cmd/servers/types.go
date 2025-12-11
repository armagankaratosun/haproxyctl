// Package servers provides commands to manage HAProxy servers.
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
package servers

import (
	"errors"
	"fmt"
	"haproxyctl/internal"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const (
	serverArgsTwo       = 2
	defaultServerWeight = 100
)

// ServerConfig represents the full server object in HAProxy Data Plane API.
type ServerConfig struct {
	APIVersion string `yaml:"apiVersion,omitempty"`
	Kind       string `yaml:"kind,omitempty"`

	Name    string `json:"name" yaml:"name"`
	Address string `json:"address" yaml:"address"`
	Port    int    `json:"port" yaml:"port"`
	Weight  int    `json:"weight,omitempty" yaml:"weight,omitempty"`
	SSL     bool   `json:"ssl,omitempty" yaml:"ssl,omitempty"`

	// Backend/Parent are used client-side to determine the parent backend
	// section (path parameter) but are not part of the v3 server object.
	Backend string `yaml:"backend,omitempty"`
	Parent  string `yaml:"parent,omitempty"`
}

// serverPayload is the subset of ServerConfig that is sent to the
// HAProxy Data Plane API, with ssl encoded as the v3 enum.
type serverPayload struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Port    int    `json:"port"`
	Weight  int    `json:"weight,omitempty"`
	SSL     string `json:"ssl,omitempty"`
}

// toPayload converts a ServerConfig into the wire-format structure
// expected by the v3 Data Plane API.
func (s ServerConfig) toPayload() serverPayload {
	payload := serverPayload{
		Name:    s.Name,
		Address: s.Address,
		Port:    s.Port,
		Weight:  s.Weight,
	}
	if s.SSL {
		payload.SSL = "enabled"
	}
	return payload
}

// NormalizeParent ensures compatibility between `parent` and `backend`.
func (s *ServerConfig) NormalizeParent() error {
	if s.Parent == "" && s.Backend != "" {
		s.Parent = s.Backend
	}
	if s.Parent == "" {
		return errors.New("server must specify parent (backend)")
	}
	return nil
}

// LoadFromFile loads a server configuration from a YAML file.
func (s *ServerConfig) LoadFromFile(filepath string) error {
	data, err := internal.LoadYAMLFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to load server configuration file: %w", err)
	}
	return yaml.Unmarshal(data, s)
}

// LoadFromFlags loads server data from CLI flags.
func (s *ServerConfig) LoadFromFlags(cmd *cobra.Command, backendName, serverName string) {
	s.Name = serverName
	s.Backend = backendName

	s.Address = internal.GetFlagString(cmd, "address")
	s.Port = internal.GetFlagInt(cmd, "port")
	s.Weight = internal.GetFlagInt(cmd, "weight")
	s.SSL = internal.GetFlagBool(cmd, "ssl")
}

// Validate performs basic validation on the ServerConfig.
func (s *ServerConfig) Validate() error {
	if s.Name == "" {
		return errors.New("server name is required")
	}
	if s.Parent == "" && s.Backend == "" {
		return errors.New("backend (parent) is required")
	}
	if s.Address == "" {
		return errors.New("server address is required")
	}
	if s.Port == 0 {
		return errors.New("server port is required")
	}
	return nil
}
