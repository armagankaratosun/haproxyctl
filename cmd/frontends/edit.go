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
	"bytes"
	"fmt"
	"haproxyctl/internal"
	"log"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// EditFrontendsCmd represents "edit frontends <name>".
var EditFrontendsCmd = &cobra.Command{
	Use:     "frontends <frontend_name>",
	Aliases: []string{"frontend"},
	Short:   "Edit a frontend definition in your editor",
	Args:    cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		frontendName := args[0]
		if err := editFrontend(frontendName); err != nil {
			log.Fatalf("Edit failed: %v", err)
		}
	},
}

func editFrontend(frontendName string) error {
	cfgVer, err := internal.GetConfigurationVersion()
	if err != nil {
		return fmt.Errorf("failed to fetch HAProxy configuration version: %w", err)
	}

	rawFrontend, err := internal.GetResource(
		"/services/haproxy/configuration/frontends/" + frontendName,
	)
	if err != nil {
		return fmt.Errorf("failed to fetch frontend %q: %w", frontendName, err)
	}

	rawBinds, err := internal.GetResourceList(
		"/services/haproxy/configuration/frontends/" + frontendName + "/binds",
	)
	if err != nil {
		// Treat missing/empty binds as non-fatal
		log.Printf("warning: failed to fetch binds for frontend %q: %v", frontendName, err)
	}

	// Build a manifest-style object: frontendWithBinds.
	var manifest frontendWithBinds
	manifest.APIVersion = "haproxyctl/v1"
	manifest.Kind = "Frontend"

	populateFrontendConfigFromMap(&manifest.frontendConfig, rawFrontend)

	for _, bind := range rawBinds {
		bc := mapBindFromAPI(bind)
		if bc.Address != "" && bc.Port != 0 {
			manifest.Binds = append(manifest.Binds, bc)
		}
	}

	origYAML, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal frontend manifest to YAML: %w", err)
	}

	tmpFile, err := internal.WriteTempYAML("haproxyctl-frontend-"+frontendName+"-", manifest)
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

	if bytes.Equal(bytes.TrimSpace(origYAML), bytes.TrimSpace(editedYAML)) {
		if _, err := fmt.Fprintln(os.Stdout, "No changes made; exiting without update."); err != nil {
			log.Printf("warning: failed to write no-change message for frontend %q: %v", frontendName, err)
		}
		return nil
	}

	var edited frontendWithBinds
	if err := yaml.Unmarshal(editedYAML, &edited); err != nil {
		return fmt.Errorf("failed to parse edited YAML: %w", err)
	}

	if edited.Name != frontendName {
		return fmt.Errorf("cannot rename frontend via edit (got %q, expected %q)", edited.Name, frontendName)
	}

	if err := edited.Validate(); err != nil {
		return fmt.Errorf("invalid frontend configuration: %w", err)
	}

	payload := edited.ToPayload()

	_, err = internal.SendRequest(
		"PUT",
		"/services/haproxy/configuration/frontends/"+frontendName,
		map[string]string{"version": strconv.Itoa(cfgVer)},
		payload,
	)
	if err != nil {
		return fmt.Errorf("failed to update frontend %q: %w", frontendName, err)
	}

	if err := applyBindDiff(frontendName, manifest.Binds, edited.Binds); err != nil {
		return fmt.Errorf("failed to apply bind changes for frontend %q: %w", frontendName, err)
	}

	if _, err := fmt.Fprintf(os.Stdout, "Frontend %q updated.\n", frontendName); err != nil {
		log.Printf("warning: failed to write frontend updated message: %v", err)
	}
	return nil
}

// mapBindFromAPI converts a generic bind object into a BindConfig suitable
// for inclusion in frontendWithBinds manifests. It keeps the underlying
// bind name (if any) for update/delete operations but does not expose it
// in the YAML, where binds are treated as anonymous sub-resources.
func mapBindFromAPI(obj map[string]interface{}) BindConfig {
	var b BindConfig

	if v, ok := obj["name"].(string); ok {
		b.Name = v
	}
	if v, ok := obj["address"].(string); ok {
		b.Address = v
	}
	if p, ok := getIntField(obj, "port"); ok {
		b.Port = p
	}
	if v, ok := obj["ssl"].(string); ok && v == "enabled" {
		b.SSL = true
	}

	return b
}

// applyBindDiff reconciles the bind list for a frontend based on the
// original and edited manifests. Since bind names are not exposed in
// the manifest, identity is determined by address+port; when a match
// is found, the original name is preserved for update/delete calls.
func applyBindDiff(frontendName string, before, after []BindConfig) error {
	beforeByKey := make(map[string]BindConfig, len(before))
	for _, b := range before {
		key := fmt.Sprintf("%s:%d", b.Address, b.Port)
		beforeByKey[key] = b
	}

	afterByKey := make(map[string]BindConfig, len(after))
	for _, b := range after {
		key := fmt.Sprintf("%s:%d", b.Address, b.Port)
		afterByKey[key] = b
	}

	// Deletes: present before, missing after.
	for key, oldB := range beforeByKey {
		if _, ok := afterByKey[key]; !ok {
			// If we don't have a name, we can't address this bind via the API.
			if oldB.Name == "" {
				log.Printf("warning: cannot delete bind %q on frontend %q: missing underlying name", key, frontendName)
				continue
			}
			if err := deleteBind(frontendName, oldB.Name); err != nil {
				return err
			}
		}
	}

	// Creates and updates.
	for key, newB := range afterByKey {
		oldB, existed := beforeByKey[key]
		if !existed {
			// New bind (no existing name needed).
			if err := createBind(frontendName, newB); err != nil {
				return err
			}
			continue
		}

		// Preserve the underlying name for updates.
		newB.Name = oldB.Name

		if bindConfigEqual(oldB, newB) {
			continue
		}

		if newB.Name == "" {
			log.Printf("warning: cannot update bind %q on frontend %q: missing underlying name", key, frontendName)
			continue
		}

		if err := updateBind(frontendName, newB); err != nil {
			return err
		}
	}

	return nil
}

// bindConfigEqual compares the fields of two BindConfig objects that
// matter to the Data Plane API (address, port, ssl), ignoring the
// internal Name used only for addressing.
func bindConfigEqual(a, b BindConfig) bool {
	return a.Address == b.Address &&
		a.Port == b.Port &&
		a.SSL == b.SSL
}
