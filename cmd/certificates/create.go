// Package certificates provides commands to manage HAProxy SSL certificates.
package certificates

import (
	"errors"
	"fmt"
	"io"
	"os"

	"haproxyctl/internal"

	"github.com/spf13/cobra"
)

// CreateCertificatesCmd represents "create certificates".
var CreateCertificatesCmd = &cobra.Command{
	Use:   "certificates <name>",
	Short: "Upload an SSL certificate bundle to HAProxy",
	Long: `Upload an SSL certificate (and optional CA chain) into the HAProxy Data Plane API storage.

You can either provide a single PEM bundle via --pem, or separate cert/key
files (and an optional CA chain) via --cert/--key/--ca-file. In all cases,
haproxyctl sends a single PEM bundle over HTTPS to the Data Plane API.

Examples:
  # Combined PEM (key + cert + chain) on disk
  haproxyctl create certificates mycert --pem /etc/haproxy/certs/mycert.pem

  # Separate cert and key, combined in memory
  haproxyctl create certificates mycert \
    --cert /etc/ssl/certs/mycert.crt \
    --key /etc/ssl/private/mycert.key \
    --ca-file /etc/ssl/certs/myca-chain.pem

  # PEM from stdin (e.g. decrypted by an external tool)
  sops -d mycert.pem.enc | haproxyctl create certificates mycert --pem -`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		pemPath := internal.GetFlagString(cmd, "pem")
		certPath := internal.GetFlagString(cmd, "cert")
		keyPath := internal.GetFlagString(cmd, "key")
		caPath := internal.GetFlagString(cmd, "ca-file")

		outputFormat := internal.GetFlagString(cmd, "output")
		dryRun := internal.GetFlagBool(cmd, "dry-run")

		fullPEM, from, err := buildCertificatePEM(pemPath, certPath, keyPath, caPath)
		if err != nil {
			return err
		}

		if outputFormat != "" || dryRun {
			// For preview/dry-run, only show metadata about the upload,
			// never the PEM contents themselves.
			info := map[string]interface{}{
				"name":   name,
				"source": from,
			}
			if outputFormat == "" {
				outputFormat = internal.OutputFormatYAML
			}
			internal.FormatOutput(info, outputFormat)
			if dryRun {
				internal.PrintDryRun()
			}
			return nil
		}

		if err := internal.UploadSSLCertificateWithContext(cmd.Context(), name, fullPEM); err != nil {
			return internal.FormatAPIError("Certificate", name, "create", err)
		}

		internal.PrintStatus("Certificate", name, internal.ActionCreated)
		return nil
	},
}

func init() {
	CreateCertificatesCmd.Flags().String("pem", "", "PEM bundle (key + cert + optional chain) file path, or '-' for stdin")
	CreateCertificatesCmd.Flags().String("cert", "", "Certificate file path (required with --key when --pem is not set)")
	CreateCertificatesCmd.Flags().String("key", "", "Private key file path (required with --cert when --pem is not set)")
	CreateCertificatesCmd.Flags().String("ca-file", "", "Optional CA/chain PEM file path to append")

	CreateCertificatesCmd.Flags().StringP("output", "o", "", "Output format: yaml or json")
	CreateCertificatesCmd.Flags().Bool("dry-run", false, "Preview certificate upload without sending it")
}

// buildCertificatePEM assembles a single PEM bundle from either a ready-made
// PEM file (--pem) or separate cert/key/ca files. It returns the PEM bytes
// along with a human-readable description of the source.
func buildCertificatePEM(pemPath, certPath, keyPath, caPath string) ([]byte, string, error) {
	hasPEM := pemPath != ""
	hasCertOrKey := certPath != "" || keyPath != "" || caPath != ""

	if hasPEM && hasCertOrKey {
		return nil, "", errors.New("cannot combine --pem with --cert/--key/--ca-file; choose one style of input")
	}

	if hasPEM {
		data, err := readMaybeStdin(pemPath)
		if err != nil {
			return nil, "", fmt.Errorf("failed to read PEM bundle %q: %w", pemPath, err)
		}
		desc := "pem:" + pemPath
		return data, desc, nil
	}

	if certPath == "" || keyPath == "" {
		return nil, "", errors.New("either --pem or both --cert and --key must be provided")
	}

	keyBytes, err := readMaybeStdin(keyPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read key %q: %w", keyPath, err)
	}

	certBytes, err := readMaybeStdin(certPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read cert %q: %w", certPath, err)
	}

	var caBytes []byte
	if caPath != "" {
		caBytes, err = readMaybeStdin(caPath)
		if err != nil {
			return nil, "", fmt.Errorf("failed to read ca-file %q: %w", caPath, err)
		}
	}

	full := append([]byte{}, keyBytes...)
	if len(full) > 0 && len(certBytes) > 0 && full[len(full)-1] != '\n' {
		full = append(full, '\n')
	}
	full = append(full, certBytes...)
	if len(caBytes) > 0 {
		if len(full) > 0 && full[len(full)-1] != '\n' {
			full = append(full, '\n')
		}
		full = append(full, caBytes...)
	}

	desc := fmt.Sprintf("cert:%s,key:%s", certPath, keyPath)
	if caPath != "" {
		desc += ",ca-file:" + caPath
	}
	return full, desc, nil
}

// readMaybeStdin reads from a file path or stdin when path is "-".
func readMaybeStdin(path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	data, err := os.ReadFile(path) //nolint:gosec // path comes from explicit CLI input
	if err != nil {
		return nil, err
	}
	return data, nil
}
