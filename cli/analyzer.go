package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:  "analyze",
	Long: "Analyze the extracted queries and generate a cache plan",
	RunE: func(cmd *cobra.Command, args []string) error {
		sqlFile := cmd.Flag("sql").Value.String()
		outFile := cmd.Flag("out").Value.String()

		// read sql file
		queries, err := readQueriesFromFile(sqlFile)
	},
}

func init() {
	analyzeCmd.Flags().StringP("sql", "s", "extracted.sql", "File containing extracted queries")
	analyzeCmd.Flags().StringP("out", "o", "isuc.yaml", "Destination file that cache plan will be written to")
	rootCmd.AddCommand(analyzeCmd)
}

func readQueriesFromFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
}
