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
	"context"
	"fmt"
	"os"
	"os/exec"

	"gopkg.in/yaml.v2"
)

// WriteTempYAML marshals v to YAML and writes it to a temporary file,
// returning the file path.
func WriteTempYAML(prefix string, v interface{}) (string, error) {
	file, err := os.CreateTemp("", prefix+"*.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			// Best-effort close; log and continue.
			fmt.Fprintf(os.Stderr, "warning: failed to close temp file: %v\n", cerr)
		}
	}()

	data, err := yaml.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if _, err := file.Write(data); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	return file.Name(), nil
}

// OpenInEditor opens the given path in the user's preferred editor,
// blocking until the editor exits.
func OpenInEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}

	// The editor command is intentionally user-controlled ($EDITOR/$VISUAL),
	// which is a common pattern for CLIs that open an interactive editor.
	cmd := exec.CommandContext(context.Background(), editor, path) //nolint:gosec // launching user-selected editor is expected behavior
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run editor %q: %w", editor, err)
	}
	return nil
}
