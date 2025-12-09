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
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

// Config holds HAProxy Data Plane API connection details.
// It mirrors what the `haproxyctl auth` command writes to
// ~/.config/haproxyctl/config.json:
//
//	{
//	  "api_base_url": "https://example:5555",
//	  "username": "admin",
//	  "password": "secret"
//	}
type Config struct {
	APIBaseURL string `json:"api_base_url"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

// Default config file path.
var configFilePath string

func init() {
	usr, _ := user.Current()
	configFilePath = filepath.Join(usr.HomeDir, ".config", "haproxyctl", "config.json")
}

// LoadConfig loads API configuration from ~/.config/haproxyctl/config.json.
func LoadConfig() (Config, error) {
	var cfg Config
	file, err := os.ReadFile(configFilePath)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config file: %w", err)
	}
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to parse config file: %w", err)
	}
	return cfg, nil
}
