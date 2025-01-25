package static_extractor

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/traP-jp/h24w-17/normalizer"
)

var sqlPattern = regexp.MustCompile(`(?i)\b(SELECT|INSERT|UPDATE|DELETE)\b`)
var replacePattern = regexp.MustCompile(`\s+`)

type ExtractedQuery struct {
	file    string
	pos     int
	content string
}

func ExtractQueryFromFile(path string, root string) ([]*ExtractedQuery, error) {
	fs := token.NewFileSet()
	node, err := parser.ParseFile(fs, path, nil, parser.AllErrors)
	if err != nil {
		return nil, fmt.Errorf("error parsing file: %v", err)
	}

	// 結果を収集
	var results []*ExtractedQuery
	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		// 文字列リテラルを抽出
		if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			// SQLクエリらしき文字列を抽出
			value := strings.Trim(lit.Value, "\"`")
			value = strings.ReplaceAll(value, "\n", " ")
			value = replacePattern.ReplaceAllString(value, " ")
			value = normalizer.NormalizeQuery(value)
			pos := lit.Pos()
			if sqlPattern.MatchString(value) {
				pos := fs.Position(pos)
				relativePath, err := filepath.Rel(root, path)
				if err != nil {
					return false
				}
				results = append(results, &ExtractedQuery{
					file:    relativePath,
					pos:     pos.Line,
					content: value,
				})
			}
			return false
		}
		return true
	})

	return results, nil
}
