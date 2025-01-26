package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/traP-jp/h24w-17/analyzer"
	"github.com/traP-jp/h24w-17/domains"
)

var analyzeCmd = &cobra.Command{
	Use:  "analyze",
	Long: "Analyze the extracted queries and generate a cache plan",
	RunE: func(cmd *cobra.Command, args []string) error {
		sqlFile := cmd.Flag("sql").Value.String()
		schemasFile := cmd.Flag("schema").Value.String()
		outFile := cmd.Flag("out").Value.String()

		// read sql file
		queries, err := readQueriesFromFile(sqlFile)
		if err != nil {
			return fmt.Errorf("failed to read queries from file: %w", err)
		}

		// load table schemas
		schemas, err := readSchemasFromFile(schemasFile)
		if err != nil {
			return fmt.Errorf("failed to read schemas from file: %w", err)
		}

		// analyze queries
		cachePlan, err := analyzer.AnalyzeQueries(queries, schemas)
		if err != nil {
			fmt.Printf("warnings:\n%v\n", err)
		}

		// write cache plan to file
		file, err := os.Create(outFile)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer file.Close()
		writer := bufio.NewWriter(file)
		err = domains.SaveCachePlan(writer, &cachePlan)
		if err != nil {
			return fmt.Errorf("failed to save cache plan: %w", err)
		}
		writer.Flush()

		return nil
	},
}

func init() {
	analyzeCmd.Flags().StringP("sql", "s", "extracted.sql", "File containing extracted queries")
	analyzeCmd.Flags().StringP("schema", "t", "schema.sql", "File containing table schemas")
	analyzeCmd.Flags().StringP("out", "o", "isuc.yaml", "Destination file that cache plan will be written to")
	rootCmd.AddCommand(analyzeCmd)
}

var commentRegex = regexp.MustCompile(`(?m)--.*$`)

func readQueriesFromFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	content := string(data)
	content = commentRegex.ReplaceAllString(content, "")
	queries := strings.Split(content, ";")
	return queries, nil
}

func readSchemasFromFile(path string) ([]domains.TableSchema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	content := string(data)
	content = commentRegex.ReplaceAllString(content, "")
	schemas, err := domains.LoadTableSchema(content)
	if err != nil {
		return nil, fmt.Errorf("failed to load table schemas: %w", err)
	}
	return schemas, nil
}
