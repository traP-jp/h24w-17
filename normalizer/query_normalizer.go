package normalizer

import (
	"regexp"
	"strings"
)

var spaceRegex = regexp.MustCompile(`\s+`)
var insertRegex = regexp.MustCompile(`INSERT INTO (\w+)\s*\(`)
var inRegex = regexp.MustCompile(`IN\s*\((\?,\s*)+\?\)`)
var valuesRegex = regexp.MustCompile(`VALUES\s*\((\?,\s*)+\?\)`)

func NormalizeQuery(query string) string {
	// remove spaces
	query = strings.ReplaceAll(query, "\r", " ")
	query = strings.ReplaceAll(query, "\n", " ")
	query = strings.ReplaceAll(query, "\t", " ")
	query = spaceRegex.ReplaceAllString(query, " ")
	query = strings.TrimSpace(query)

	// remove backquotes
	query = strings.ReplaceAll(query, "`", "")

	// INSERT INTO table(... -> INSERT INTO table (...
	query = insertRegex.ReplaceAllString(query, "INSERT INTO $1 (")

	// IN (?, ?, ?) -> IN (?)
	query = inRegex.ReplaceAllString(query, "IN (?)")

	// VALUES (?, ?, ?) -> VALUES (?)
	query = valuesRegex.ReplaceAllString(query, "VALUES (?)")

	// add semicolon
	if !strings.HasSuffix(query, ";") {
		query += ";"
	}

	return query
}
