package cmd

import (
	"log"

	"haproxyctl/cmd/acls"
	"haproxyctl/cmd/backends"
	"haproxyctl/cmd/configuration"
	"haproxyctl/cmd/frontends"
	"haproxyctl/cmd/servers"
	"haproxyctl/internal"

	"github.com/spf13/cobra"
)

// getCmd represents the "get" command
var getCmd = &cobra.Command{
	Use:   "get <resource> [name]",
	Short: "Retrieve information from HAProxy",
	Long: `Fetch details about HAProxy configuration, including backends, frontends, servers, and ACLs.

Examples:
  haproxyctl get configuration version -o json
  haproxyctl get backend mybackend -o yaml
  haproxyctl get frontends -o json
  haproxyctl get acl myfrontend -o yaml`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		normalizedResource, err := internal.NormalizeResource(args[0])
		if err != nil {
			log.Fatalf("Unknown resource type: %s", args[0])
		}

		// Extract optional resource name argument
		resourceName := internal.ExtractOptionalArg(args)

		// Call the corresponding Cobra command directly
		var subCmd *cobra.Command

		switch normalizedResource {
		case "backends":
			subCmd = backends.GetBackendsCmd
		case "frontends":
			subCmd = frontends.GetFrontendsCmd
		case "servers":
			subCmd = servers.GetServersCmd
		case "acls":
			subCmd = acls.GetACLsCmd
		case "configuration":
			subCmd = configuration.GetConfigurationCmd
		default:
			log.Fatalf("Unsupported resource type: %s", normalizedResource)
		}

		// Execute the subcommand's Run function with the extracted arguments
		subCmd.Run(cmd, append([]string{resourceName}, args[1:]...))
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Add subcommands
	getCmd.AddCommand(acls.GetACLsCmd)
	getCmd.AddCommand(backends.GetBackendsCmd)
	getCmd.AddCommand(configuration.GetConfigurationCmd)
	getCmd.AddCommand(frontends.GetFrontendsCmd)
	getCmd.AddCommand(servers.GetServersCmd)

	getCmd.PersistentFlags().StringP("output", "o", "", "Output format: yaml or json (default: table)")
}
