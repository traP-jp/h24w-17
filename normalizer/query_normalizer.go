package normalizer

import "regexp"

// IN (?, ?, ?) -> IN (?)
// VALUES (?, ?, ?) -> VALUES (?)

var inRegex = regexp.MustCompile(`IN\s+\((\?,\s*)+\?\)`)
var valuesRegex = regexp.MustCompile(`VALUES\s+\((\?,\s*)+\?\)`)

func NormalizeQuery(query string) string {
	query = inRegex.ReplaceAllString(query, "IN (?)")
	query = valuesRegex.ReplaceAllString(query, "VALUES (?)")
	return query
}
