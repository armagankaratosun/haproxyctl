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
	"errors"
	"fmt"
	"strings"
)

const (
	apiVersionV1   = "haproxyctl/v1"
	userlistKind   = "Userlist"
	userlistGoKind = "Userlist"
)

// UserlistManifest represents a high-level, manifest-style view of a userlist.
type UserlistManifest struct {
	APIVersion string          `json:"apiVersion" yaml:"apiVersion"`
	Kind       string          `json:"kind" yaml:"kind"`
	Name       string          `json:"name" yaml:"name"`
	Users      []UserManifest  `json:"users,omitempty" yaml:"users,omitempty"`
	Groups     []GroupManifest `json:"groups,omitempty" yaml:"groups,omitempty"`
}

// UserManifest represents a single user in a userlist manifest.
type UserManifest struct {
	Name     string   `json:"name" yaml:"name"`
	Password string   `json:"password" yaml:"password"`
	Groups   []string `json:"groups,omitempty" yaml:"groups,omitempty"`
}

// GroupManifest represents a single group in a userlist manifest.
type GroupManifest struct {
	Name  string   `json:"name" yaml:"name"`
	Users []string `json:"users,omitempty" yaml:"users,omitempty"`
}

// manifestFromAPI converts a raw API userlist object into a manifest.
func manifestFromAPI(obj map[string]interface{}) (*UserlistManifest, error) {
	name, _ := obj["name"].(string)
	if name == "" {
		return nil, errors.New("userlist object is missing name")
	}

	manifest := &UserlistManifest{
		APIVersion: apiVersionV1,
		Kind:       userlistKind,
		Name:       name,
	}

	// Users
	if rawUsers, ok := obj["users"]; ok {
		if userMap, ok := rawUsers.(map[string]interface{}); ok {
			for _, v := range userMap {
				user, ok := v.(map[string]interface{})
				if !ok {
					continue
				}
				username, _ := user["username"].(string)
				if username == "" {
					continue
				}
				password, _ := user["password"].(string)
				groupsStr, _ := user["groups"].(string)
				var groups []string
				if groupsStr != "" {
					for _, g := range strings.Split(groupsStr, ",") {
						g = strings.TrimSpace(g)
						if g != "" {
							groups = append(groups, g)
						}
					}
				}
				manifest.Users = append(manifest.Users, UserManifest{
					Name:     username,
					Password: password,
					Groups:   groups,
				})
			}
		}
	}

	// Groups
	if rawGroups, ok := obj["groups"]; ok {
		if groupMap, ok := rawGroups.(map[string]interface{}); ok {
			for _, v := range groupMap {
				group, ok := v.(map[string]interface{})
				if !ok {
					continue
				}
				name, _ := group["name"].(string)
				if name == "" {
					continue
				}
				usersStr, _ := group["users"].(string)
				var users []string
				if usersStr != "" {
					for _, u := range strings.Split(usersStr, ",") {
						u = strings.TrimSpace(u)
						if u != "" {
							users = append(users, u)
						}
					}
				}
				manifest.Groups = append(manifest.Groups, GroupManifest{
					Name:  name,
					Users: users,
				})
			}
		}
	}

	return manifest, nil
}

// toAPIPayload converts a manifest into the wire-format map expected by the
// Data Plane API userlists endpoints.
func (m *UserlistManifest) toAPIPayload() (map[string]interface{}, error) {
	if m.Name == "" {
		return nil, errors.New("userlist manifest is missing name")
	}
	if m.Kind != "" && m.Kind != userlistKind {
		return nil, fmt.Errorf("kind must be %q", userlistKind)
	}
	if m.APIVersion != "" && m.APIVersion != apiVersionV1 {
		return nil, fmt.Errorf("apiVersion must be %q", apiVersionV1)
	}

	payload := map[string]interface{}{
		"name": m.Name,
	}

	if len(m.Users) > 0 {
		users := make(map[string]interface{})
		for _, u := range m.Users {
			if u.Name == "" {
				return nil, errors.New("user entry is missing name")
			}
			user := map[string]interface{}{
				"username":        u.Name,
				"password":        u.Password,
				"secure_password": true,
			}
			if len(u.Groups) > 0 {
				user["groups"] = strings.Join(u.Groups, ",")
			}
			users[u.Name] = user
		}
		payload["users"] = users
	}

	if len(m.Groups) > 0 {
		groups := make(map[string]interface{})
		for _, g := range m.Groups {
			if g.Name == "" {
				return nil, errors.New("group entry is missing name")
			}
			group := map[string]interface{}{
				"name": g.Name,
			}
			if len(g.Users) > 0 {
				group["users"] = strings.Join(g.Users, ",")
			}
			groups[g.Name] = group
		}
		payload["groups"] = groups
	}

	return payload, nil
}
