package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/h24w-17/extractor"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:       "query-extractor",
	Long:      "Statistically analyze the codebase and extract SQL queries",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"path"},
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		out, err := cmd.Flags().GetString("out")
		if err != nil {
			return fmt.Errorf("error getting out flag: %v", err)
		}

		valid := extractor.IsValidDir(path)
		if !valid {
			return fmt.Errorf("invalid directory: %s", path)
		}
		files, err := extractor.ListAllGoFiles(path)
		if err != nil {
			return fmt.Errorf("error listing go files: %v", err)
		}

		fmt.Printf("found %d go files\n", len(files))
		extractedQueries := []*extractor.ExtractedQuery{}
		for _, file := range files {
			relativePath, err := filepath.Rel(path, file)
			if err != nil {
				return fmt.Errorf("error getting relative path: %v", err)
			}
			queries, err := extractor.ExtractQueryFromFile(file, path)
			if err != nil {
				return fmt.Errorf("❌ %s: error while extracting: %v", relativePath, err)
			}
			fmt.Printf("✅ %s: %d queries extracted\n", relativePath, len(queries))
			extractedQueries = append(extractedQueries, queries...)
		}

		err = extractor.WriteQueriesToFile(out, extractedQueries)
		if err != nil {
			return fmt.Errorf("error writing queries to file: %v", err)
		}

		fmt.Printf("queries written to %s\n", out)

		return nil
	},
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("out", "o", "extracted.sql", "Destination file that extracted queries will be written to")
}
