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

// Package transactions provides commands to inspect HAProxy configuration transactions.
package transactions

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"haproxyctl/internal"

	"github.com/spf13/cobra"
)

// GetTransactionsCmd represents "get transactions".
var GetTransactionsCmd = &cobra.Command{
	Use:     "transactions [id]",
	Aliases: []string{"transaction"},
	Short:   "List HAProxy transactions or fetch details of a specific transaction",
	Long: `Retrieve HAProxy configuration transactions from the Data Plane API.

Examples:
  haproxyctl get transactions
  haproxyctl get transactions --status in_progress
  haproxyctl get transactions 273e3385-2d0c-4fb1-aa27-93cbb31ff203`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var id string
		if len(args) > 0 {
			id = args[0]
		}
		getTransactions(cmd, id)
	},
}

// getTransactions handles fetching transactions (list or single item).
func getTransactions(cmd *cobra.Command, id string) {
	outputFormat := internal.GetFlagString(cmd, "output")
	if outputFormat == "" {
		outputFormat = "table"
	}

	statusFilter := internal.GetFlagString(cmd, "status")

	var data interface{}
	var err error

	if id == "" {
		data, err = getTransactionsListFromAPI(cmd, statusFilter)
	} else {
		data, err = getTransactionFromAPI(cmd, id)
	}

	if err != nil {
		if id != "" && internal.IsNotFoundError(err) {
			_, _ = fmt.Fprintln(os.Stdout, internal.ResourceID("Transaction", id)+" not found")
			return
		}
		log.Fatalf("Failed to fetch transaction(s): %v", err)
	}

	internal.FormatOutput(data, outputFormat)
}

// getTransactionsListFromAPI fetches the list of transactions, optionally filtered by status.
func getTransactionsListFromAPI(cmd *cobra.Command, status string) ([]map[string]interface{}, error) {
	query := map[string]string{}
	if status != "" {
		query["status"] = status
	}

	data, err := internal.SendRequestWithContext(cmd.Context(), "GET", "/services/haproxy/transactions", query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}

	var list []map[string]interface{}
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("failed to parse transactions list response: %w", err)
	}

	// Ensure deterministic ordering by transaction ID when present.
	internal.SortByStringField(list, "id")

	return list, nil
}

// getTransactionFromAPI fetches a single transaction by id.
func getTransactionFromAPI(cmd *cobra.Command, id string) (map[string]interface{}, error) {
	data, err := internal.SendRequestWithContext(cmd.Context(), "GET", "/services/haproxy/transactions/"+id, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction %q: %w", id, err)
	}

	var tx map[string]interface{}
	if err := json.Unmarshal(data, &tx); err != nil {
		return nil, fmt.Errorf("failed to parse transaction response: %w", err)
	}

	return tx, nil
}

func init() {
	GetTransactionsCmd.Flags().StringP("output", "o", "", "Output format: table, yaml, or json")
	GetTransactionsCmd.Flags().String("status", "", "Filter transactions by status (failed or in_progress)")
}
