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
	"encoding/json"
	"fmt"
	"haproxyctl/internal"
	"log"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

// GetUserlistsCmd represents "get userlists".
var GetUserlistsCmd = &cobra.Command{
	Use:     "userlists [name]",
	Aliases: []string{"userlist"},
	Short:   "List HAProxy userlists or fetch details of a specific userlist",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) > 0 {
			name = args[0]
		}
		getUserlists(cmd, name)
	},
}

func getUserlists(cmd *cobra.Command, name string) {
	outputFormat := internal.GetFlagString(cmd, "output")

	if name == "" {
		list, err := internal.GetResourceList("/services/haproxy/configuration/userlists")
		if err != nil {
			log.Fatalf("Failed to fetch userlists: %v", err)
		}

		if outputFormat == "" {
			outputFormat = "table"
		}

		internal.SortByStringField(list, "name")
		internal.FormatOutput(list, outputFormat)
		return
	}

	manifest, err := getUserlistManifest(cmd, name)
	if err != nil {
		if internal.IsNotFoundError(err) {
			_, _ = fmt.Fprintln(os.Stdout, internal.ResourceID("Userlist", name)+" not found")
			return
		}
		log.Fatalf("Failed to fetch userlist %q: %v", name, err)
	}

	if outputFormat == "" {
		row := map[string]interface{}{
			"name":   manifest.Name,
			"users":  len(manifest.Users),
			"groups": len(manifest.Groups),
		}
		internal.FormatOutput(row, "table")
		return
	}

	internal.FormatOutput(manifest, outputFormat)
}

// getUserlistManifest fetches a single userlist (with full_section=true) and
// converts it into a manifest.
func getUserlistManifest(cmd *cobra.Command, name string) (*UserlistManifest, error) {
	escaped := url.PathEscape(name)
	endpoint := "/services/haproxy/configuration/userlists/" + escaped

	raw, err := internal.SendRequestWithContext(cmd.Context(), "GET", endpoint, map[string]string{"full_section": "true"}, nil)
	if err != nil {
		return nil, err
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, fmt.Errorf("failed to parse userlist response: %w", err)
	}

	return manifestFromAPI(obj)
}
