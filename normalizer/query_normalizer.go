package normalizer

import "regexp"

// IN (?, ?, ?) -> IN (?)
// VALUES (?, ?, ?) -> VALUES (?)

var inRegex = regexp.MustCompile(`IN\s+\((\?,\s*)+\?\)`)
var valuesRegex = regexp.MustCompile(`VALUES\s+\((\?,\s*)+\?\)`)

func NormalizeQuery(query string) (string, error) {
	query = inRegex.ReplaceAllString(query, "IN (?)")
	query = valuesRegex.ReplaceAllString(query, "VALUES (?)")
	return query, nil
}
