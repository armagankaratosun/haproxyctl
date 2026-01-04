// Package certificates provides commands to manage HAProxy SSL certificates.
package certificates

import (
	"encoding/json"
	"fmt"
	"haproxyctl/internal"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// GetCertificatesCmd represents "get certificates".
var GetCertificatesCmd = &cobra.Command{
	Use:     "certificates [name]",
	Aliases: []string{"certificate"},
	Short:   "List SSL certificates or show details for one",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) > 0 {
			name = args[0]
		}
		getCertificates(cmd, name)
	},
}

func init() {
	GetCertificatesCmd.Flags().StringP("output", "o", "", "Output format: yaml or json")
}

func getCertificates(cmd *cobra.Command, name string) {
	outputFormat := internal.GetFlagString(cmd, "output")

	endpoint := "/services/haproxy/storage/ssl_certificates"

	data, err := internal.SendRequestWithContext(cmd.Context(), "GET", endpoint, nil, nil)
	if err != nil {
		log.Fatalf("Failed to fetch certificate(s): %v", err)
	}

	var list []map[string]interface{}
	if err := json.Unmarshal(data, &list); err != nil {
		log.Fatalf("Failed to parse certificates list response: %v\nResponse: %s", err, string(data))
	}

	if name == "" {
		// List view.
		internal.SortByStringField(list, "storage_name")
		rows := make([]interface{}, 0, len(list))
		for _, m := range list {
			rows = append(rows, m)
		}
		internal.FormatOutput(rows, outputFormat)
		return
	}

	// Named view: find by storage_name.
	var found map[string]interface{}
	for _, m := range list {
		if v, ok := m["storage_name"].(string); ok && v == name {
			found = m
			break
		}
	}

	if found == nil {
		_, _ = fmt.Fprintln(os.Stdout, internal.ResourceID("Certificate", name)+" not found")
		return
	}

	internal.FormatOutput(found, outputFormat)
}
