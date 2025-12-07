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

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Interactively configure your Data Plane API authentication",
	Long: `Set up your HAProxy Data Plane API authentication configuration
This command will prompt you for the following values:
api_base_url, username and password via interactive prompts.  
These values get written to:
  $HOME/.config/haproxyctl/config.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// validator that disallows empty strings
		validateNonEmpty := func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("value cannot be empty")
			}
			return nil
		}

		// 1) API Base URL (free-form)
		apiPrompt := promptui.Prompt{
			Label:    "API Base URL",
			Default:  viper.GetString("api_base_url"),
			Validate: validateNonEmpty,
		}
		api_base_url, err := apiPrompt.Run()
		if err != nil {
			return fmt.Errorf("prompt failed for api base url: %w", err)
		}

		// 2) Username (free-form)
		usernamePrompt := promptui.Prompt{
			Label:    "Username",
			Default:  viper.GetString("username"),
			Validate: validateNonEmpty,
		}
		username, err := usernamePrompt.Run()
		if err != nil {
			return fmt.Errorf("prompt failed for username: %w", err)
		}
		// 3) Password (hidden input)
		passwordPrompt := promptui.Prompt{
			Label:    "Password",
			Validate: validateNonEmpty,
			Mask:     '*', // mask input with asterisks
		}
		password, err := passwordPrompt.Run()
		if err != nil {
			return fmt.Errorf("prompt failed for password: %w", err)
		}

		// 3.2) Create config directory if it doesn't exist
		configDir := filepath.Join(os.Getenv("HOME"), ".config", "haproxyctl")
		configFile := filepath.Join(configDir, "config.json")
		// viper setup
		viper.SetConfigFile(configFile)
		viper.SetConfigType("json")

		// Ensure the config directory exists
		// 3.1) Check if the config directory exists, if not, create
		if err := os.MkdirAll(configDir, 0o700); err != nil {
			return fmt.Errorf("cannot create config dir %q: %w", configDir, err)
		}
		// 3.2) make sure the file itself exists
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			// If the file does not exist, create it with appropriate permissions
			if err := os.WriteFile(configFile, []byte{}, 0o600); err != nil {
				return fmt.Errorf("cannot create config file %q: %w", configFile, err)
			}
		}
		// 3.4) Set the values
		// Note: viper will create the config file if it doesn't exist
		// and write the values to it.
		// This is a good place to set defaults if needed.
		// e.g., viper.SetDefault("api_base_url", "https://api.example.com")
		//       viper.SetDefault("username", "admin")
		//       viper.SetDefault("password", "password")
		viper.Set("api_base_url", api_base_url)
		viper.Set("username", username)
		viper.Set("password", password)
		// 3.4) Overwrite the config file
		// This will write the config to the specified file.
		// If the file already exists, it will be overwritten.
		if err := viper.WriteConfigAs(configFile); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}
		// 4) Print the configuration
		fmt.Printf("Configuration saved to %s\n", configFile)

		return nil

	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
