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

// Package userlists provides commands to manage HAProxy userlists, users, and groups.
package userlists

import (
	"fmt"
	"haproxyctl/internal"

	"gopkg.in/yaml.v2"
)

// CreateUserlistFromFile creates a userlist from a manifest YAML payload.
func CreateUserlistFromFile(data []byte) error {
	var manifest UserlistManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("failed to parse userlist manifest: %w", err)
	}

	if manifest.APIVersion != "" && manifest.APIVersion != apiVersionV1 {
		return fmt.Errorf("unsupported apiVersion %q (expected %s)", manifest.APIVersion, apiVersionV1)
	}
	if manifest.Kind != "" && manifest.Kind != userlistKind {
		return fmt.Errorf("invalid kind %q, expected %q", manifest.Kind, userlistKind)
	}

	payload, err := manifest.toAPIPayload()
	if err != nil {
		return err
	}

	_, err = internal.SendRequest("POST", "/services/haproxy/configuration/userlists", nil, payload)
	if err != nil {
		return internal.FormatAPIError("Userlist", manifest.Name, "create", err)
	}

	internal.PrintStatus("Userlist", manifest.Name, internal.ActionCreated)
	return nil
}
