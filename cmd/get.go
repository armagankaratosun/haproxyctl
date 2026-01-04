package cmd

import (
	"haproxyctl/cmd/acls"
	"haproxyctl/cmd/backends"
	"haproxyctl/cmd/certificates"
	"haproxyctl/cmd/configuration"
	"haproxyctl/cmd/frontends"
	"haproxyctl/cmd/reloads"
	"haproxyctl/cmd/servers"
	"haproxyctl/cmd/stats"
	"haproxyctl/cmd/transactions"
	"haproxyctl/cmd/userlists"

	"github.com/spf13/cobra"
)

// getCmd represents the "get" command.
var getCmd = &cobra.Command{
	Use:   "get <resource> [name]",
	Short: "Retrieve information from HAProxy",
	Long: `Fetch details about HAProxy configuration, including backends, frontends, servers, and ACLs.

Examples:
  haproxyctl get configuration version -o json
  haproxyctl get backend mybackend -o yaml
  haproxyctl get frontends -o json
  haproxyctl get acl myfrontend -o yaml`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// If no subcommand was provided, show help.
		return cmd.Help()
	},
}

// init initializes the get command and its subcommands.
// It adds the get command to the root command and sets up its subcommands.
// It also sets up the persistent flags for output formatting.

func init() {
	rootCmd.AddCommand(getCmd)

	// Add subcommands.
	getCmd.AddCommand(acls.GetACLsCmd)
	getCmd.AddCommand(backends.GetBackendsCmd)
	getCmd.AddCommand(certificates.GetCertificatesCmd)
	getCmd.AddCommand(configuration.GetConfigurationCmd)
	getCmd.AddCommand(frontends.GetFrontendsCmd)
	getCmd.AddCommand(reloads.GetReloadsCmd)
	getCmd.AddCommand(stats.GetStatsCmd)
	getCmd.AddCommand(transactions.GetTransactionsCmd)
	getCmd.AddCommand(userlists.GetUserlistsCmd)
	getCmd.AddCommand(servers.GetServersCmd)

	getCmd.PersistentFlags().StringP("output", "o", "", "Output format: yaml or json (default: table)")
}
