package normalizer

import (
	"regexp"
	"strings"
)

// IN (?, ?, ?) -> IN (?)
// VALUES (?, ?, ?) -> VALUES (?)

var inRegex = regexp.MustCompile(`IN\s+\((\?,\s*)+\?\)`)
var valuesRegex = regexp.MustCompile(`VALUES\s+\((\?,\s*)+\?\)`)

func NormalizeQuery(query string) string {
	query = strings.TrimSpace(query)
	query = inRegex.ReplaceAllString(query, "IN (?)")
	query = valuesRegex.ReplaceAllString(query, "VALUES (?)")
	if !strings.HasSuffix(query, ";") {
		query += ";"
	}
	return query
}
