package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	h24w17 "github.com/traP-jp/h24w-17"
)

var generateCmd = &cobra.Command{
	Use:       "generate",
	Long:      "Generate a driver from the cache plan and table schema",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"path"},
	RunE: func(cmd *cobra.Command, args []string) error {
		plan, err := cmd.Flags().GetString("plan")
		if err != nil {
			return fmt.Errorf("error getting plan flag: %v", err)
		}
		schema, err := cmd.Flags().GetString("schema")
		if err != nil {
			return fmt.Errorf("error getting schema flag: %v", err)
		}
		distPath := args[0]

		planContent, err := readFile(plan)
		if err != nil {
			return fmt.Errorf("error reading plan file: %v", err)
		}
		schemaContent, err := readFile(schema)
		if err != nil {
			return fmt.Errorf("error reading schema file: %v", err)
		}

		g := h24w17.NewGenerator(planContent, schemaContent)
		g.Generate(distPath)

		return nil
	},
}

func init() {
	generateCmd.Flags().StringP("plan", "p", "isuc.yaml", "File containing the cache plan")
	generateCmd.Flags().StringP("schema", "s", "schema.sql", "File containing the table schema")
	rootCmd.AddCommand(generateCmd)
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}
	return string(data), nil
}
