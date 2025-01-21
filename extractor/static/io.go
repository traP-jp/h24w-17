package static_extractor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func IsValidDir(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func ListAllGoFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return []string{}, fmt.Errorf("error walking directory: %v", err)
	}
	return files, nil
}

func WriteQueriesToFile(out string, queries []*ExtractedQuery) error {
	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer f.Close()

	for _, query := range queries {
		_, err := f.WriteString(query.String() + "\n")
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
	}

	return nil
}

func (q *ExtractedQuery) String() string {
	return fmt.Sprintf("-- %s:%d\n%s", q.file, q.pos, q.content)
}
