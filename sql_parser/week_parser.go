package sql_parser

import (
	"fmt"
	"regexp"
)

var selectStmtRegex = regexp.MustCompile("^SELECT\\s+`?(?P<columns>.+)`?\\s+FROM\\s+`?(?P<table>\\w+)`?")
var updateStmtRegex = regexp.MustCompile("^UPDATE\\s+`?(?P<table>\\w+)`?")
var deleteStmtRegex = regexp.MustCompile("^DELETE\\s+FROM\\s+`?(?P<table>\\w+)`?")
var insertStmtRegex = regexp.MustCompile("^INSERT\\s+INTO\\s+`?(?P<table>\\w+)`?")

func ParseSQLWeekly(query string) (SQLNode, error) {
	if match := matchRegex(selectStmtRegex, query); match != nil {
		return SelectStmtNode{Table: TableNode{Name: match["table"]}}, nil
	}
	if match := matchRegex(updateStmtRegex, query); match != nil {
		return UpdateStmtNode{Table: TableNode{Name: match["table"]}}, nil
	}
	if match := matchRegex(deleteStmtRegex, query); match != nil {
		return DeleteStmtNode{Table: TableNode{Name: match["table"]}}, nil
	}
	if match := matchRegex(insertStmtRegex, query); match != nil {
		return InsertStmtNode{Table: TableNode{Name: match["table"]}}, nil
	}

	return nil, fmt.Errorf("failed to parse query: \"%s\"", query)
}

func matchRegex(re *regexp.Regexp, query string) map[string]string {
	match := re.FindStringSubmatch(query)
	if match == nil {
		return nil
	}

	result := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return result
}
