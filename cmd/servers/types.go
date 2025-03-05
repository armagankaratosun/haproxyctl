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
	"fmt"
	"haproxyctl/utils"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// ServerConfig represents the full server object in HAProxy Data Plane API
type ServerConfig struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`

	Name    string `json:"name" yaml:"name"`
	Address string `json:"address" yaml:"address"`
	Port    int    `json:"port" yaml:"port"`
	Weight  int    `json:"weight,omitempty" yaml:"weight,omitempty"`
	SSL     bool   `json:"ssl,omitempty" yaml:"ssl,omitempty"`

	Backend string `json:"backend,omitempty" yaml:"backend,omitempty"`
	Parent  string `json:"parent,omitempty" yaml:"parent,omitempty"`
}

// NormalizeParent ensures compatibility between `parent` and `backend`
func (s *ServerConfig) NormalizeParent() error {
	if s.Parent == "" && s.Backend != "" {
		s.Parent = s.Backend
	}
	if s.Parent == "" {
		return fmt.Errorf("server must specify parent (backend)")
	}
	return nil
}

// LoadFromFile loads a server configuration from a YAML file
func (s *ServerConfig) LoadFromFile(filepath string) error {
	data, err := utils.LoadYAMLFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to load server configuration file: %w", err)
	}
	return yaml.Unmarshal(data, s)
}

// LoadFromFlags loads server data from CLI flags
func (s *ServerConfig) LoadFromFlags(cmd *cobra.Command, backendName, serverName string) {
	s.Name = serverName
	s.Backend = backendName

	s.Address = utils.GetFlagString(cmd, "address")
	s.Port = utils.GetFlagInt(cmd, "port")
	s.Weight = utils.GetFlagInt(cmd, "weight")
	s.SSL = utils.GetFlagBool(cmd, "ssl")
}

// Validate performs basic validation on the ServerConfig
func (s *ServerConfig) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("server name is required")
	}
	if s.Backend == "" {
		return fmt.Errorf("backend (parent) is required")
	}
	if s.Address == "" {
		return fmt.Errorf("server address is required")
	}
	if s.Port == 0 {
		return fmt.Errorf("server port is required")
	}
	return nil
}
