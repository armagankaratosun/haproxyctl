// Package backends provides commands to manage HAProxy backends.
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
	"bytes"
	"fmt"
	"haproxyctl/cmd/servers"
	"haproxyctl/internal"
	"log"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// EditBackendsCmd represents "edit backends <name>".
var EditBackendsCmd = &cobra.Command{
	Use:     "backends <backend_name>",
	Aliases: []string{"backend"},
	Short:   "Edit a backend definition in your editor",
	Args:    cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		backendName := args[0]
		if err := editBackend(backendName); err != nil {
			log.Fatalf("Edit failed: %v", err)
		}
	},
}

func editBackend(backendName string) error {
	cfgVer, err := internal.GetConfigurationVersion()
	if err != nil {
		return fmt.Errorf("failed to fetch HAProxy configuration version: %w", err)
	}

	// Fetch the current backend and its servers.
	rawBackend, err := internal.GetResource(
		"/services/haproxy/configuration/backends/" + backendName,
	)
	if err != nil {
		return fmt.Errorf("failed to fetch backend %q: %w", backendName, err)
	}

	rawServers, err := internal.GetResourceList(
		"/services/haproxy/configuration/backends/" + backendName + "/servers",
	)
	if err != nil {
		// Treat missing/empty servers as non-fatal
		log.Printf("warning: failed to fetch servers for backend %q: %v", backendName, err)
	}

	// Build manifest-style object: backendWithServers
	var manifest backendWithServers
	manifest.APIVersion = "haproxyctl/v1"
	manifest.Kind = backendKind

	populateBackendConfigFromMap(&manifest.backendConfig, rawBackend)

	for _, srv := range rawServers {
		sc := mapServerFromAPI(backendName, srv)
		if sc.Name != "" && sc.Address != "" && sc.Port != 0 {
			manifest.Servers = append(manifest.Servers, sc)
		}
	}

	origYAML, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal backend manifest to YAML: %w", err)
	}

	tmpFile, err := internal.WriteTempYAML("haproxyctl-backend-"+backendName+"-", manifest)
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

	// tmpFile is created via os.CreateTemp in WriteTempYAML and lives in
	// the system temp directory, so this read is safe.
	editedYAML, err := os.ReadFile(tmpFile) //nolint:gosec // tmpFile is controlled by this process
	if err != nil {
		return fmt.Errorf("failed to read edited file: %w", err)
	}

	// If unchanged, exit quietly
	if bytes.Equal(bytes.TrimSpace(origYAML), bytes.TrimSpace(editedYAML)) {
		if _, err := fmt.Fprintln(os.Stdout, "No changes made; exiting without update."); err != nil {
			log.Printf("warning: failed to write no-change message for backend %q: %v", backendName, err)
		}
		return nil
	}

	var edited backendWithServers
	if err := yaml.Unmarshal(editedYAML, &edited); err != nil {
		return fmt.Errorf("failed to parse edited YAML: %w", err)
	}

	if edited.Name != backendName {
		return fmt.Errorf("cannot rename backend via edit (got %q, expected %q)", edited.Name, backendName)
	}

	if err := edited.Validate(); err != nil {
		return fmt.Errorf("invalid backend configuration: %w", err)
	}

	payload := edited.toPayload()

	_, err = internal.SendRequest(
		"PUT",
		"/services/haproxy/configuration/backends/"+backendName,
		map[string]string{"version": strconv.Itoa(cfgVer)},
		payload,
	)
	if err != nil {
		return fmt.Errorf("failed to update backend %q: %w", backendName, err)
	}

	// Apply server changes based on the edited manifest.
	if err := applyServerDiff(backendName, manifest.Servers, edited.Servers); err != nil {
		return fmt.Errorf("failed to apply server changes for backend %q: %w", backendName, err)
	}

	internal.PrintStatus("Backend", backendName, internal.ActionConfigured)
	return nil
}

// populateBackendConfigFromMap maps a generic API backend object into
// the strongly-typed backendConfig used by manifests.
func populateBackendConfigFromMap(cfg *backendConfig, obj map[string]interface{}) {
	if v, ok := obj["name"].(string); ok {
		cfg.Name = v
	}
	if v, ok := obj["mode"].(string); ok {
		cfg.Mode = v
	}
	if m, ok := obj["balance"].(map[string]interface{}); ok {
		cfg.Balance = toStringMap(m)
	}
	if m, ok := obj["default_server"].(map[string]interface{}); ok {
		cfg.DefaultServer = m
	}
	if m, ok := obj["forwardfor"].(map[string]interface{}); ok {
		cfg.ForwardFor = toStringMap(m)
	}

	// Timeouts come back as integer milliseconds; render them as
	// human-readable strings for the manifest.
	if ms, ok := getIntField(obj, "timeout_client"); ok {
		cfg.TimeoutClient = internal.FormatMillisAsDuration(ms)
	}
	if ms, ok := getIntField(obj, "timeout_http_keep_alive"); ok {
		cfg.TimeoutHTTPKeepAlive = internal.FormatMillisAsDuration(ms)
	}
	if ms, ok := getIntField(obj, "timeout_http_request"); ok {
		cfg.TimeoutHTTPRequest = internal.FormatMillisAsDuration(ms)
	}
	if ms, ok := getIntField(obj, "timeout_queue"); ok {
		cfg.TimeoutQueue = internal.FormatMillisAsDuration(ms)
	}
	if ms, ok := getIntField(obj, "timeout_server"); ok {
		cfg.TimeoutServer = internal.FormatMillisAsDuration(ms)
	}
	if ms, ok := getIntField(obj, "timeout_server_fin"); ok {
		cfg.TimeoutServerFin = internal.FormatMillisAsDuration(ms)
	}

	if v, ok := obj["tcpka"].(string); ok && v == stateEnabled {
		cfg.TCPKA = true
	}

	if rm, ok := obj["redispatch"].(map[string]interface{}); ok {
		if v, ok := rm["enabled"].(string); ok && v == stateEnabled {
			cfg.Redispatch = true
		}
	}

	if m, ok := obj["source"].(map[string]interface{}); ok {
		cfg.Source = toStringMap(m)
	}
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

// mapServerFromAPI converts a generic server object into a ServerConfig
// suitable for inclusion in backendWithServers manifests. It treats
// "ssl" enum values as booleans and preserves the backend name.
func mapServerFromAPI(backendName string, obj map[string]interface{}) servers.ServerConfig {
	var sc servers.ServerConfig

	if v, ok := obj["name"].(string); ok {
		sc.Name = v
	}
	if v, ok := obj["address"].(string); ok {
		sc.Address = v
	}
	if ms, ok := getIntField(obj, "port"); ok {
		sc.Port = ms
	}
	if w, ok := getIntField(obj, "weight"); ok {
		sc.Weight = w
	}
	if v, ok := obj["ssl"].(string); ok && v == "enabled" {
		sc.SSL = true
	}

	sc.Backend = backendName
	return sc
}

// applyServerDiff reconciles the server list for a backend based on the
// original and edited manifests. It creates, updates, or deletes servers
// using the existing server helpers in the servers package.
func applyServerDiff(backendName string, before, after []servers.ServerConfig) error {
	beforeByName := make(map[string]servers.ServerConfig, len(before))
	for _, s := range before {
		beforeByName[s.Name] = s
	}

	afterByName := make(map[string]servers.ServerConfig, len(after))
	for _, s := range after {
		// Ensure backend is set for downstream helpers
		if s.Backend == "" {
			s.Backend = backendName
		}
		afterByName[s.Name] = s
	}

	// Deletes: present before, missing after.
	for name := range beforeByName {
		if _, ok := afterByName[name]; !ok {
			if err := servers.DeleteServer(backendName, name); err != nil {
				return err
			}
		}
	}

	// Creates and updates.
	for name, newS := range afterByName {
		oldS, existed := beforeByName[name]
		if !existed {
			if err := servers.CreateServer(newS, "", false); err != nil {
				return err
			}
			continue
		}

		if serverConfigEqual(oldS, newS) {
			continue
		}

		if err := servers.UpdateServer(newS); err != nil {
			return err
		}
	}

	return nil
}

// serverConfigEqual compares the fields of two ServerConfig objects that
// are significant to the Data Plane API (name, address, port, weight, ssl).
func serverConfigEqual(a, b servers.ServerConfig) bool {
	return a.Name == b.Name &&
		a.Address == b.Address &&
		a.Port == b.Port &&
		a.Weight == b.Weight &&
		a.SSL == b.SSL
}
