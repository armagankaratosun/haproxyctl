// Package certificates provides commands to manage HAProxy SSL certificates.
package certificates

import (
	"haproxyctl/internal"
	"log"

	"github.com/spf13/cobra"
)

// DeleteCertificatesCmd represents "delete certificates".
var DeleteCertificatesCmd = &cobra.Command{
	Use:   "certificates <name>",
	Short: "Delete an SSL certificate from HAProxy storage",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if err := deleteCertificate(cmd, name); err != nil {
			log.Fatalf("Failed to delete certificate %q: %v", name, err)
		}
	},
}

func deleteCertificate(cmd *cobra.Command, name string) error {
	endpoint := "/services/haproxy/storage/ssl_certificates/" + name
	_, err := internal.SendRequestWithContext(cmd.Context(), "DELETE", endpoint, nil, nil)
	if err != nil {
		return internal.FormatAPIError("Certificate", name, "delete", err)
	}

	internal.PrintStatus("Certificate", name, internal.ActionDeleted)
	return nil
}
