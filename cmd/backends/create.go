package backends

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"haproxyctl/utils"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// CreateBackendsCmd represents "create backend"
var CreateBackendsCmd = &cobra.Command{
	Use:   "backend <backend_name>",
	Short: "Create a new HAProxy backend",
	Long: `Creates a new HAProxy backend with optional parameters.

Examples:
  haproxyctl create backend mybackend --mode http --balance roundrobin --maxconn 50
  haproxyctl create backend another_backend --mode tcp --balance leastconn --check enabled`,
	Args: cobra.ExactArgs(1), // Requires exactly one argument: backend name
	Run: func(cmd *cobra.Command, args []string) {
		backendName := args[0]
		createBackend(backendName, cmd)
	},
}

// createBackend creates a new HAProxy backend with the provided name and options
func createBackend(backendName string, cmd *cobra.Command) {
	if backendName == "" {
		log.Fatal("Backend name is required. Usage: haproxyctl create backend <backend_name> --mode <mode> --balance <algorithm> ...")
	}

	// Get flag values
	mode, _ := cmd.Flags().GetString("mode")
	balance, _ := cmd.Flags().GetString("balance")
	alpn, _ := cmd.Flags().GetString("alpn")
	check, _ := cmd.Flags().GetString("check")
	checkAlpn, _ := cmd.Flags().GetString("check-alpn")
	maxconn, _ := cmd.Flags().GetInt("maxconn")
	weight, _ := cmd.Flags().GetInt("weight")
	outputFormat, _ := cmd.Flags().GetString("output") // -o yaml or -o json
	dryRun, _ := cmd.Flags().GetBool("dry-run")        // --dry-run flag

	// Prepare backend payload
	backendData := map[string]interface{}{
		"name": backendName,
		"mode": mode,
		"balance": map[string]string{
			"algorithm": balance,
		},
		"default_server": map[string]interface{}{
			"alpn":       alpn,
			"check":      check,
			"check_alpn": checkAlpn,
			"maxconn":    maxconn,
			"weight":     weight,
		},
	}

	// Handle YAML or JSON output format
	if outputFormat == "yaml" {
		yamlOutput, err := yaml.Marshal(backendData)
		if err != nil {
			log.Fatal("Failed to generate YAML:", err)
		}
		fmt.Println(string(yamlOutput))
		return
	} else if outputFormat == "json" {
		jsonOutput, err := json.MarshalIndent(backendData, "", "    ")
		if err != nil {
			log.Fatal("Failed to generate JSON:", err)
		}
		fmt.Println(string(jsonOutput))
		return
	}

	// If dry-run is enabled, print the payload and exit
	if dryRun {
		fmt.Println("Dry run mode enabled. No changes will be made.")
		jsonOutput, err := json.MarshalIndent(backendData, "", "    ")
		if err != nil {
			log.Fatal("Failed to format JSON:", err)
		}
		fmt.Println(string(jsonOutput))
		return
	}

	// Fetch HAProxy configuration version
	data, err := utils.SendRequest("GET", "/services/haproxy/configuration/version", nil, nil)
	if err != nil {
		log.Fatal("Failed to fetch HAProxy configuration version:", err)
	}

	versionStr := strings.TrimSpace(string(data)) // Trim newline and spaces
	versionInt, err := strconv.Atoi(versionStr)
	if err != nil {
		log.Fatal("Failed to parse version as an integer:", err)
	}

	// Send request to create backend
	_, err = utils.SendRequest("POST", "/services/haproxy/configuration/backends",
		map[string]string{"version": strconv.Itoa(versionInt)},
		backendData,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Backend '%s' created successfully.\n", backendName)
}

func init() {
	// Register flags for backend creation
	CreateBackendsCmd.Flags().String("mode", "http", "Backend mode (default: http)")
	CreateBackendsCmd.Flags().String("balance", "roundrobin", "Load balancing algorithm (default: roundrobin)")
	CreateBackendsCmd.Flags().String("alpn", "h2", "Application-Layer Protocol Negotiation (default: h2)")
	CreateBackendsCmd.Flags().String("check", "enabled", "Enable or disable server health checks (default: enabled)")
	CreateBackendsCmd.Flags().String("check-alpn", "h2", "Check ALPN protocol (default: h2)")
	CreateBackendsCmd.Flags().Int("maxconn", 100, "Max connections per server (default: 100)")
	CreateBackendsCmd.Flags().Int("weight", 100, "Server weight (default: 100)")
	CreateBackendsCmd.Flags().StringP("output", "o", "", "Output format: yaml or json")
	CreateBackendsCmd.Flags().Bool("dry-run", false, "Simulate the request without making changes")
}
