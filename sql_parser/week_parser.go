package sql_parser

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/traP-jp/isuc/domains"
)

var selectStmtRegex = regexp.MustCompile("^SELECT\\s+`?(?P<columns>.+)`?\\s+FROM\\s+`?(?P<table>\\w+)`?")
var updateStmtRegex = regexp.MustCompile("^UPDATE\\s+`?(?P<table>\\w+)`?")
var deleteStmtRegex = regexp.MustCompile("^DELETE\\s+FROM\\s+`?(?P<table>\\w+)`?")
var insertStmtRegex = regexp.MustCompile("^INSERT\\s+INTO\\s+`?(?P<table>\\w+)`?")

func ParseSQLWeekly(query string, schemas []domains.TableSchema) (SQLNode, error) {
	if match := matchRegex(selectStmtRegex, query); match != nil {
		return SelectStmtNode{Table: TableNode{Name: match["table"]}}, nil
	}
	if match := matchRegex(updateStmtRegex, query); match != nil {
		schema := domains.TableSchema{}
		for _, s := range schemas {
			if s.TableName == match["table"] {
				schema = s
				break
			}
		}
		if schema.TableName == "" {
			return nil, fmt.Errorf("table not found: %s", match["table"])
		}
		sets := make([]UpdateSetNode, 0, len(schema.Columns))
		// ensure the order of columns
		keys := make([]string, 0, len(schema.Columns))
		for key := range schema.Columns {
			keys = append(keys, key)
		}
		slices.Sort(keys)
		for _, column := range keys {
			sets = append(sets, UpdateSetNode{Column: ColumnNode{Name: schema.Columns[column].ColumnName}, Value: PlaceholderNode{}})
		}
		return UpdateStmtNode{Table: TableNode{Name: match["table"]}, Sets: UpdateSetsNode{Sets: sets}}, nil
	}
	if match := matchRegex(deleteStmtRegex, query); match != nil {
		return DeleteStmtNode{Table: TableNode{Name: match["table"]}}, nil
	}
	if match := matchRegex(insertStmtRegex, query); match != nil {
		return InsertStmtNode{Table: TableNode{Name: match["table"]}}, nil
	}

	// \n -> \\n
	query = fmt.Sprintf("%q", query)
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
