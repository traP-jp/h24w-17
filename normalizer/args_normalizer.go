package normalizer

import (
	"fmt"

	"github.com/traP-jp/h24w-17/sql_parser"
)

type ExtraArg struct {
	Column string
	Value  interface{}
}

type NormalizedArgs struct {
	Query     string
	ExtraSets []ExtraArg
	ExtraArgs []ExtraArg
}

func NormalizeArgs(query string) (NormalizedArgs, error) {
	parsed, err := sql_parser.ParseSQL(query)
	if err != nil {
		return NormalizedArgs{Query: query}, fmt.Errorf("failed to parse sql: %w", err)
	}
	extracted, err := extractExtraArgs(parsed)
	if err != nil {
		return NormalizedArgs{Query: query}, fmt.Errorf("failed to transform sql: %w", err)
	}
	return NormalizedArgs{
		Query:     extracted.node.String(),
		ExtraSets: extracted.sets,
		ExtraArgs: extracted.args,
	}, nil
}

type extractResult struct {
	node sql_parser.SQLNode
	sets []ExtraArg
	args []ExtraArg
}

func extractExtraArgs(node sql_parser.SQLNode) (out extractResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			out.node = node
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	// TODO: generalize select, update, and delete statements
	switch n := node.(type) {
	case sql_parser.SelectStmtNode:
		args := make([]ExtraArg, 0)
		if n.Conditions != nil {
			extracted, err := extractExtraArgs(*n.Conditions)
			if err != nil {
				return extractResult{node: n}, fmt.Errorf("SelectStmtNode: %w", err)
			}
			conditions := extracted.node.(sql_parser.ConditionsNode)
			n.Conditions = &conditions
			args = append(args, extracted.args...)
		}
		if n.Limit != nil {
			extracted, err := extractExtraArgs(*n.Limit)
			if err != nil {
				return extractResult{node: n}, fmt.Errorf("SelectStmtNode: %w", err)
			}
			limit := extracted.node.(sql_parser.LimitNode)
			n.Limit = &limit
			args = append(args, extracted.args...)
		}
		if n.Offset != nil {
			extracted, err := extractExtraArgs(*n.Offset)
			if err != nil {
				return extractResult{node: n}, fmt.Errorf("SelectStmtNode: %w", err)
			}
			offset := extracted.node.(sql_parser.OffsetNode)
			n.Offset = &offset
			args = append(args, extracted.args...)
		}
		return extractResult{node: n, args: args}, nil
	case sql_parser.UpdateStmtNode:
		sets := make([]ExtraArg, 0)
		args := make([]ExtraArg, 0)
		for i, s := range n.Sets.Sets {
			switch v := s.Value.(type) {
			case sql_parser.StringNode:
				sets = append(sets, ExtraArg{Column: s.Column.Name, Value: v.Value})
				n.Sets.Sets[i].Value = sql_parser.PlaceholderNode{}
			case sql_parser.NumberNode:
				sets = append(sets, ExtraArg{Column: s.Column.Name, Value: v.Value})
				n.Sets.Sets[i].Value = sql_parser.PlaceholderNode{}
			}
		}
		if n.Conditions != nil {
			extracted, err := extractExtraArgs(*n.Conditions)
			if err != nil {
				return extractResult{node: n}, fmt.Errorf("SelectStmtNode: %w", err)
			}
			conditions := extracted.node.(sql_parser.ConditionsNode)
			n.Conditions = &conditions
			args = append(args, extracted.args...)
		}
		if n.Limit != nil {
			extracted, err := extractExtraArgs(*n.Limit)
			if err != nil {
				return extractResult{node: n}, fmt.Errorf("SelectStmtNode: %w", err)
			}
			limit := extracted.node.(sql_parser.LimitNode)
			n.Limit = &limit
			args = append(args, extracted.args...)
		}
		if n.Offset != nil {
			extracted, err := extractExtraArgs(*n.Offset)
			if err != nil {
				return extractResult{node: n}, fmt.Errorf("SelectStmtNode: %w", err)
			}
			offset := extracted.node.(sql_parser.OffsetNode)
			n.Offset = &offset
			args = append(args, extracted.args...)
		}
		return extractResult{node: n, sets: sets, args: args}, nil
	case sql_parser.DeleteStmtNode:
		if n.Conditions == nil {
			return extractResult{node: n}, nil
		}
		extracted, err := extractExtraArgs(*n.Conditions)
		if err != nil {
			return extractResult{node: n}, fmt.Errorf("DeleteStmtNode: %w", err)
		}
		conditions := extracted.node.(sql_parser.ConditionsNode)
		n.Conditions = &conditions
		return extractResult{node: n, args: extracted.args}, nil
	case sql_parser.ConditionsNode:
		conditions := make([]sql_parser.ConditionNode, 0, len(n.Conditions))
		args := make([]ExtraArg, 0, len(n.Conditions))
		for i, c := range n.Conditions {
			extracted, err := extractExtraArgs(c)
			if err != nil {
				return extractResult{node: n}, fmt.Errorf("ConditionsNode[%d]: %w", i, err)
			}
			conditions = append(conditions, extracted.node.(sql_parser.ConditionNode))
			args = append(args, extracted.args...)
		}
		n.Conditions = conditions
		return extractResult{node: n, args: args}, nil
	case sql_parser.ConditionNode:
		args := make([]ExtraArg, 0)
		if n.Operator.Operator != sql_parser.Operator_IN && n.Operator.Operator != sql_parser.Operator_EQ {
			return extractResult{node: n, args: args}, nil
		}
		switch val := n.Value.(type) {
		case sql_parser.ValuesNode:
			transformed := make([]sql_parser.SQLNode, 0, len(val.Values))
			for _, value := range val.Values {
				switch value := value.(type) {
				case sql_parser.StringNode:
					args = append(args, ExtraArg{Column: n.Column.Name, Value: value.Value})
					transformed = append(transformed, sql_parser.PlaceholderNode{})
					continue
				case sql_parser.NumberNode:
					args = append(args, ExtraArg{Column: n.Column.Name, Value: value.Value})
					transformed = append(transformed, sql_parser.PlaceholderNode{})
					continue
				}
				transformed = append(transformed, value)
			}
			n.Value = sql_parser.ValuesNode{Values: transformed}
		case sql_parser.StringNode:
			args = append(args, ExtraArg{Column: n.Column.Name, Value: val.Value})
			n.Value = sql_parser.PlaceholderNode{}
		case sql_parser.NumberNode:
			args = append(args, ExtraArg{Column: n.Column.Name, Value: val.Value})
			n.Value = sql_parser.PlaceholderNode{}
		}
		return extractResult{node: n, args: args}, nil
	case sql_parser.LimitNode:
		if limit, ok := n.Limit.(sql_parser.NumberNode); ok {
			n.Limit = sql_parser.PlaceholderNode{}
			return extractResult{node: n, args: []ExtraArg{{Column: "LIMIT()", Value: limit.Value}}}, nil
		}
		return extractResult{node: n}, nil
	case sql_parser.OffsetNode:
		if offset, ok := n.Offset.(sql_parser.NumberNode); ok {
			n.Offset = sql_parser.PlaceholderNode{}
			return extractResult{node: n, args: []ExtraArg{{Column: "OFFSET()", Value: offset.Value}}}, nil
		}
		return extractResult{node: n}, nil
	default:
		return extractResult{node: n}, nil
	}
}
