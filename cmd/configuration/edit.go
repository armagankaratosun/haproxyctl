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

// Package configuration provides commands to manage HAProxy global configuration.
package configuration

import (
	"bytes"
	"errors"
	"fmt"
	"haproxyctl/internal"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// EditConfigurationCmd groups configuration edit subcommands under
// "haproxyctl edit configuration".
var EditConfigurationCmd = &cobra.Command{
	Use:   "configuration",
	Short: "Edit HAProxy configuration sections",
	RunE: func(cmd *cobra.Command, _ []string) error {
		// Show help if no subcommand is provided.
		return cmd.Help()
	},
}

// EditGlobalsCmd represents "edit configuration globals".
var EditGlobalsCmd = &cobra.Command{
	Use:     "globals",
	Aliases: []string{"global"},
	Short:   "Edit HAProxy global configuration in your editor",
	Args:    cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		if err := editSection(
			"/services/haproxy/configuration/global",
			"Global",
			"haproxyctl-global-",
			func(obj map[string]interface{}) interface{} { return mapGlobalFromAPI(obj) },
			func(version int, cfg interface{}) error {
				g, ok := cfg.(GlobalConfig)
				if !ok {
					return fmt.Errorf("expected GlobalConfig, got %T", cfg)
				}
				return putGlobal(version, g)
			},
		); err != nil {
			log.Fatalf("Edit globals failed: %v", err)
		}
	},
}

// EditDefaultsCmd represents "edit configuration defaults".
var EditDefaultsCmd = &cobra.Command{
	Use:   "defaults",
	Short: "Edit HAProxy defaults configuration in your editor",
	Args:  cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		if err := editSection(
			"/services/haproxy/configuration/defaults",
			"Defaults",
			"haproxyctl-defaults-",
			func(obj map[string]interface{}) interface{} { return mapDefaultsFromAPI(obj) },
			func(version int, cfg interface{}) error {
				d, ok := cfg.(DefaultsConfig)
				if !ok {
					return fmt.Errorf("expected DefaultsConfig, got %T", cfg)
				}
				return putDefaults(version, d)
			},
		); err != nil {
			log.Fatalf("Edit defaults failed: %v", err)
		}
	},
}

func init() {
	EditConfigurationCmd.AddCommand(EditGlobalsCmd)
	EditConfigurationCmd.AddCommand(EditDefaultsCmd)
}

func editSection(
	getEndpoint string,
	kind string,
	tmpPrefix string,
	mapFromAPI func(map[string]interface{}) interface{},
	putFn func(int, interface{}) error,
) error {
	version, err := internal.GetConfigurationVersion()
	if err != nil {
		return fmt.Errorf("failed to fetch HAProxy configuration version: %w", err)
	}

	var obj map[string]interface{}
	if kind == "Defaults" {
		list, lerr := internal.GetResourceList(getEndpoint)
		if lerr != nil {
			return fmt.Errorf("failed to fetch %s configuration: %w", strings.ToLower(kind), lerr)
		}
		if len(list) == 0 {
			obj = map[string]interface{}{}
		} else {
			obj = list[0]
		}
	} else {
		obj, err = internal.GetResource(getEndpoint)
		if err != nil {
			return fmt.Errorf("failed to fetch %s configuration: %w", strings.ToLower(kind), err)
		}
	}

	manifest := mapFromAPI(obj)

	origYAML, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal %s configuration to YAML: %w", strings.ToLower(kind), err)
	}

	tmpFile, err := internal.WriteTempYAML(tmpPrefix, manifest)
	if err != nil {
		return err
	}
	defer func() {
		if rmErr := os.Remove(tmpFile); rmErr != nil {
			log.Printf("warning: failed to remove temp file %q: %v", tmpFile, rmErr)
		}
	}()

	if err := internal.OpenInEditor(tmpFile); err != nil {
		return err
	}

	editedYAML, err := os.ReadFile(tmpFile) //nolint:gosec // tmpFile is controlled by this process
	if err != nil {
		return fmt.Errorf("failed to read edited file: %w", err)
	}

	if bytes.Equal(bytes.TrimSpace(origYAML), bytes.TrimSpace(editedYAML)) {
		internal.PrintStatus(kind, "config", internal.ActionUnchanged)
		return nil
	}

	var edited interface{}
	switch kind {
	case "Global":
		var g GlobalConfig
		if err := yaml.Unmarshal(editedYAML, &g); err != nil {
			return fmt.Errorf("failed to parse edited global YAML: %w", err)
		}
		edited = g
	case "Defaults":
		var d DefaultsConfig
		if err := yaml.Unmarshal(editedYAML, &d); err != nil {
			return fmt.Errorf("failed to parse edited defaults YAML: %w", err)
		}
		edited = d
	default:
		return fmt.Errorf("unsupported kind %q", kind)
	}

	if err := putFn(version, edited); err != nil {
		return err
	}

	internal.PrintStatus(kind, "config", internal.ActionConfigured)
	return nil
}

func putGlobal(version int, cfg GlobalConfig) error {
	payload := map[string]interface{}{
		"daemon":  cfg.Daemon,
		"nbproc":  cfg.Nbproc,
		"maxconn": cfg.Maxconn,
	}

	if cfg.Log != "" {
		payload["log"] = cfg.Log
	}
	if cfg.LogSendHost != "" {
		payload["log_send_hostname"] = cfg.LogSendHost
	}
	if cfg.StatsSocket != "" {
		payload["stats_socket"] = cfg.StatsSocket
	}
	if cfg.StatsTimeout != "" {
		payload["stats_timeout"] = cfg.StatsTimeout
	}
	if cfg.SpreadChecks != 0 {
		payload["spread_checks"] = cfg.SpreadChecks
	}

	_, err := internal.SendRequest(
		"PUT",
		"/services/haproxy/configuration/global",
		map[string]string{"version": strconv.Itoa(version)},
		payload,
	)
	if err != nil {
		return fmt.Errorf("failed to update global configuration: %w", err)
	}
	return nil
}

func putDefaults(version int, cfg DefaultsConfig) error {
	if cfg.Name == "" {
		return errors.New("defaults name is required to update configuration")
	}

	payload := map[string]interface{}{}

	if cfg.Mode != "" {
		payload["mode"] = cfg.Mode
	}
	if cfg.TimeoutClient != "" {
		payload["timeout_client"] = cfg.TimeoutClient
	}
	if cfg.TimeoutServer != "" {
		payload["timeout_server"] = cfg.TimeoutServer
	}
	if cfg.TimeoutConnect != "" {
		payload["timeout_connect"] = cfg.TimeoutConnect
	}
	if cfg.TimeoutQueue != "" {
		payload["timeout_queue"] = cfg.TimeoutQueue
	}
	if cfg.TimeoutTunnel != "" {
		payload["timeout_tunnel"] = cfg.TimeoutTunnel
	}
	if cfg.Balance != "" {
		payload["balance"] = cfg.Balance
	}
	if cfg.Log != "" {
		payload["log"] = cfg.Log
	}

	endpoint := "/services/haproxy/configuration/defaults/" + cfg.Name

	_, err := internal.SendRequest(
		"PUT",
		endpoint,
		map[string]string{"version": strconv.Itoa(version)},
		payload,
	)
	if err != nil {
		return fmt.Errorf("failed to update defaults configuration: %w", err)
	}
	return nil
}
