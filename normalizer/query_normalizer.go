package normalizer

import (
	"fmt"

	"github.com/traP-jp/h24w-17/sql_parser"
)

// IN (?, ?, ?) -> IN (?)
// VALUES (?, ?, ?) -> VALUES (?)
// select id from table -> SELECT `id` FROM `table`

func NormalizedQuery(query string) (string, error) {
	parsed, err := sql_parser.ParseSQL(query)
	if err != nil {
		return query, fmt.Errorf("failed to parse sql: %w", err)
	}
	transformed, err := transform(parsed)
	if err != nil {
		return query, fmt.Errorf("failed to transform sql: %w", err)
	}
	return transformed.String(), nil
}

func transform(node sql_parser.SQLNode) (out sql_parser.SQLNode, err error) {
	defer func() {
		if r := recover(); r != nil {
			out = node
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	switch n := node.(type) {
	case sql_parser.SelectStmtNode:
		if n.Conditions == nil {
			return n, nil
		}
		transformed, err := transform(*n.Conditions)
		if err != nil {
			return n, fmt.Errorf("failed to transform conditions: %w", err)
		}
		conditions := transformed.(sql_parser.ConditionsNode)
		n.Conditions = &conditions
		return n, nil
	case sql_parser.UpdateStmtNode:
		if n.Conditions == nil {
			return n, nil
		}
		transformed, err := transform(*n.Conditions)
		if err != nil {
			return n, fmt.Errorf("failed to transform conditions: %w", err)
		}
		conditions := transformed.(sql_parser.ConditionsNode)
		n.Conditions = &conditions
		return n, nil
	case sql_parser.InsertStmtNode:
		transformed, err := transform(n.Values)
		if err != nil {
			return n, fmt.Errorf("failed to transform values: %w", err)
		}
		n.Values = transformed.(sql_parser.ValuesNode)
		return n, nil
	case sql_parser.DeleteStmtNode:
		if n.Conditions == nil {
			return n, nil
		}
		transformed, err := transform(*n.Conditions)
		if err != nil {
			return n, fmt.Errorf("failed to transform conditions: %w", err)
		}
		conditions := transformed.(sql_parser.ConditionsNode)
		n.Conditions = &conditions
		return n, nil
	case sql_parser.ConditionsNode:
		transformed := make([]sql_parser.ConditionNode, 0, len(n.Conditions))
		for _, c := range n.Conditions {
			visited, err := transform(c)
			if err != nil {
				return n, err
			}
			transformed = append(transformed, visited.(sql_parser.ConditionNode))
		}
		n.Conditions = transformed
		return n, nil
	case sql_parser.ConditionNode:
		if n.Operator.Operator != sql_parser.Operator_IN {
			return n, nil
		}
		val := n.Value.(sql_parser.ValuesNode)
		if len(val.Values) == 0 {
			return n, fmt.Errorf("unexpected value length: %d", len(val.Values))
		}
		// look at the first one only (for simplicity)
		// false positive case: (?, 1, 2) -> (?)
		if val.Values[0].String() != "?" {
			return n, nil
		}
		val.Values = []sql_parser.SQLNode{sql_parser.PlaceholderNode{}}
		n.Value = val
		return n, nil
	case sql_parser.ValuesNode:
		if len(n.Values) == 0 {
			return n, fmt.Errorf("unexpected value length: %d", len(n.Values))
		}
		// look at the first one only (for simplicity)
		// false positive case: (?, 1, 2) -> (?)
		if n.Values[0].String() != "?" {
			return n, nil
		}
		n.Values = []sql_parser.SQLNode{sql_parser.PlaceholderNode{}}
		return n, nil
	}
	return node, nil
}
