package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	static_extractor "github.com/traP-jp/h24w-17/extractor/static"
)

var extractCmd = &cobra.Command{
	Use:       "extract",
	Long:      "Statistically analyze the codebase and extract SQL queries",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"path"},
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		out, err := cmd.Flags().GetString("out")
		if err != nil {
			return fmt.Errorf("error getting out flag: %v", err)
		}

		valid := static_extractor.IsValidDir(path)
		if !valid {
			return fmt.Errorf("invalid directory: %s", path)
		}
		files, err := static_extractor.ListAllGoFiles(path)
		if err != nil {
			return fmt.Errorf("error listing go files: %v", err)
		}

		fmt.Printf("found %d go files\n", len(files))
		extractedQueries := []*static_extractor.ExtractedQuery{}
		for _, file := range files {
			relativePath, err := filepath.Rel(path, file)
			if err != nil {
				return fmt.Errorf("error getting relative path: %v", err)
			}
			queries, err := static_extractor.ExtractQueryFromFile(file, path)
			if err != nil {
				return fmt.Errorf("❌ %s: error while extracting: %v", relativePath, err)
			}
			fmt.Printf("✅ %s: %d queries extracted\n", relativePath, len(queries))
			extractedQueries = append(extractedQueries, queries...)
		}

		err = static_extractor.WriteQueriesToFile(out, extractedQueries)
		if err != nil {
			return fmt.Errorf("error writing queries to file: %v", err)
		}

		fmt.Printf("queries written to %s\n", out)

		return nil
	},
}

func init() {
	extractCmd.Flags().StringP("out", "o", "extracted.sql", "Destination file that extracted queries will be written to")
	rootCmd.AddCommand(extractCmd)
}
