package analyzer

import (
	"fmt"

	"github.com/traP-jp/isuc/domains"
	"github.com/traP-jp/isuc/sql_parser"
)

func newQueryAnalyzer(schemas []domains.TableSchema) *queryAnalyzer {
	return &queryAnalyzer{
		schemas:          schemas,
		placeholderIndex: 0,
	}
}

type queryAnalyzer struct {
	schemas          []domains.TableSchema
	placeholderIndex int
}

func (q *queryAnalyzer) placeholder() int {
	q.placeholderIndex++
	return q.placeholderIndex - 1
}

func (q *queryAnalyzer) analyzeSelectStmt(node sql_parser.SelectStmtNode) (domains.CachePlanQuery, error) {
	schema := domains.TableSchema{}
	for _, s := range q.schemas {
		if s.TableName == node.Table.Name {
			schema = s
			break
		}
	}
	if schema.TableName == "" {
		return domains.CachePlanQuery{}, fmt.Errorf("table schema not found for \"%s\"", node.Table.Name)
	}

	// parsed by week parser
	if len(node.Values.Values) == 0 {
		return domains.CachePlanQuery{
			CachePlanQueryBase: &domains.CachePlanQueryBase{
				Query: node.String(),
				Type:  domains.CachePlanQueryType_SELECT,
			},
			Select: &domains.CachePlanSelectQuery{
				Table: node.Table.Name,
				Cache: false,
			},
		}, nil
	}

	targets := q.analyzeSelectValues(node.Values, schema)
	selectErr := analyzerError{}
	conditions, err := q.analyzeConditions(node.Conditions)
	if err != nil {
		selectErr.errors = append(selectErr.errors, fmt.Errorf("failed to analyze conditions: %s", err))
	}
	orders, err := q.analyzeOrders(node.Orders)
	if err != nil {
		selectErr.errors = append(selectErr.errors, fmt.Errorf("failed to analyze orders: %s", err))
	}
	conditions = append(conditions, q.analyzeLimit(node.Limit, len(conditions))...)
	conditions = append(conditions, q.analyzeOffset(node.Offset, len(conditions))...)

	query := domains.CachePlanQuery{
		CachePlanQueryBase: &domains.CachePlanQueryBase{
			Query: node.String(),
			Type:  domains.CachePlanQueryType_SELECT,
		},
		Select: &domains.CachePlanSelectQuery{
			Table:      node.Table.Name,
			Cache:      true,
			Targets:    targets,
			Conditions: conditions,
			Orders:     orders,
		},
	}

	return query, selectErr.wrap()
}

func (q *queryAnalyzer) analyzeSelectValues(values sql_parser.SelectValuesNode, schema domains.TableSchema) []string {
	result := []string{}
	for _, value := range values.Values {
		switch v := value.(type) {
		case sql_parser.SelectValueAsteriskNode:
			for _, column := range schema.Columns {
				result = append(result, column.ColumnName)
			}
		case sql_parser.SelectValueColumnNode:
			result = append(result, v.Column.Name)
		case sql_parser.SelectValueFunctionNode:
			if v.Name == "COUNT" {
				result = append(result, "COUNT()")
			} else {
				switch arg := v.Value.(type) {
				case sql_parser.SelectValueAsteriskNode:
					for _, column := range schema.Columns {
						result = append(result, column.ColumnName)
					}
				case sql_parser.ColumnNode:
					result = append(result, arg.Name)
				}
			}
		}
	}
	return result
}

func (q *queryAnalyzer) analyzeInsertStmt(node sql_parser.InsertStmtNode) (domains.CachePlanQuery, error) {
	columns := q.analyzeColumns(node.Columns)
	return domains.CachePlanQuery{
		CachePlanQueryBase: &domains.CachePlanQueryBase{
			Query: node.String(),
			Type:  domains.CachePlanQueryType_INSERT,
		},
		Insert: &domains.CachePlanInsertQuery{
			Table:   node.Table.Name,
			Columns: columns,
		},
	}, nil
}

func (q *queryAnalyzer) analyzeColumns(values sql_parser.ColumnsNode) []string {
	result := []string{}
	for _, value := range values.Columns {
		result = append(result, value.Name)
	}
	return result
}

func (q *queryAnalyzer) analyzeUpdateStmt(node sql_parser.UpdateStmtNode) (domains.CachePlanQuery, error) {
	targets := q.analyzeUpdateSets(node.Sets)
	selectErr := analyzerError{}
	conditions, err := q.analyzeConditions(node.Conditions)
	if err != nil {
		selectErr.errors = append(selectErr.errors, fmt.Errorf("failed to analyze conditions: %s", err))
	}
	orders, err := q.analyzeOrders(node.Orders)
	if err != nil {
		selectErr.errors = append(selectErr.errors, fmt.Errorf("failed to analyze orders: %s", err))
	}
	conditions = append(conditions, q.analyzeLimit(node.Limit, len(conditions))...)
	conditions = append(conditions, q.analyzeOffset(node.Offset, len(conditions))...)

	query := domains.CachePlanQuery{
		CachePlanQueryBase: &domains.CachePlanQueryBase{
			Query: node.String(),
			Type:  domains.CachePlanQueryType_UPDATE,
		},
		Update: &domains.CachePlanUpdateQuery{
			Table:      node.Table.Name,
			Targets:    targets,
			Conditions: conditions,
			Orders:     orders,
		},
	}

	return query, selectErr.wrap()
}

func (q *queryAnalyzer) analyzeUpdateSets(sets sql_parser.UpdateSetsNode) []domains.CachePlanUpdateTarget {
	result := []domains.CachePlanUpdateTarget{}
	for _, set := range sets.Sets {
		if _, ok := set.Value.(sql_parser.PlaceholderNode); !ok {
			continue
		}
		result = append(result, domains.CachePlanUpdateTarget{
			Column:      set.Column.Name,
			Placeholder: domains.CachePlanPlaceholder{Index: q.placeholder()},
		})
	}
	return result
}

func (a *queryAnalyzer) analyzeDeleteStmt(node sql_parser.DeleteStmtNode) (domains.CachePlanQuery, error) {
	conditions, err := a.analyzeConditions(node.Conditions)
	if err != nil {
		return domains.CachePlanQuery{}, fmt.Errorf("failed to analyze conditions: %s", err)
	}
	conditions = append(conditions, a.analyzeLimit(node.Limit, len(conditions))...)
	conditions = append(conditions, a.analyzeOffset(node.Offset, len(conditions))...)

	return domains.CachePlanQuery{
		CachePlanQueryBase: &domains.CachePlanQueryBase{
			Query: node.String(),
			Type:  domains.CachePlanQueryType_DELETE,
		},
		Delete: &domains.CachePlanDeleteQuery{
			Table:      node.Table.Name,
			Conditions: conditions,
		},
	}, nil
}

func (a *queryAnalyzer) analyzeConditions(node *sql_parser.ConditionsNode) ([]domains.CachePlanCondition, error) {
	if node == nil {
		return []domains.CachePlanCondition{}, nil
	}
	conditionsErr := analyzerError{}
	conditions := []domains.CachePlanCondition{}
	for _, condition := range node.Conditions {
		// continue if the value is not ? or (?)
		if _, ok := condition.Value.(sql_parser.PlaceholderNode); !ok {
			v, ok := condition.Value.(sql_parser.ValuesNode)
			if !ok {
				continue
			}
			if _, ok := v.Values[0].(sql_parser.PlaceholderNode); !ok {
				continue
			}
		}

		op, err := a.analyzeOperator(condition.Operator)
		if err != nil {
			conditionsErr.errors = append(conditionsErr.errors, fmt.Errorf("failed to analyze operator: %s", err))
			continue
		}
		conditions = append(conditions, domains.CachePlanCondition{
			Column:      condition.Column.Name,
			Operator:    op,
			Placeholder: domains.CachePlanPlaceholder{Index: a.placeholder()},
		})

	}
	return conditions, conditionsErr.wrap()
}

func (q *queryAnalyzer) analyzeOrders(node *sql_parser.OrdersNode) ([]domains.CachePlanOrder, error) {
	if node == nil {
		return []domains.CachePlanOrder{}, nil
	}
	ordersErr := analyzerError{}
	orders := []domains.CachePlanOrder{}
	for _, order := range node.Orders {
		o, err := q.analyzeOrder(order)
		if err != nil {
			ordersErr.errors = append(ordersErr.errors, fmt.Errorf("failed to analyze order: %s", err))
			continue
		}
		orders = append(orders, o)
	}
	return orders, ordersErr.wrap()
}

func (a *queryAnalyzer) analyzeOrder(node sql_parser.OrderNode) (domains.CachePlanOrder, error) {
	order, err := a.analyzeEnum(node.Order)
	if err != nil {
		return domains.CachePlanOrder{}, fmt.Errorf("failed to analyze order enum: %s", err)
	}
	return domains.CachePlanOrder{
		Column: node.Column.Name,
		Order:  order,
	}, nil
}

func (q *queryAnalyzer) analyzeLimit(node *sql_parser.LimitNode, index int) []domains.CachePlanCondition {
	if node == nil {
		return []domains.CachePlanCondition{}
	}

	if _, ok := node.Limit.(sql_parser.PlaceholderNode); !ok {
		return []domains.CachePlanCondition{}
	}

	return []domains.CachePlanCondition{
		{
			Column:      "LIMIT()",
			Operator:    domains.CachePlanOperator_EQ,
			Placeholder: domains.CachePlanPlaceholder{Index: index},
		},
	}
}

func (q *queryAnalyzer) analyzeOffset(node *sql_parser.OffsetNode, index int) []domains.CachePlanCondition {
	if node == nil {
		return []domains.CachePlanCondition{}
	}

	if _, ok := node.Offset.(sql_parser.PlaceholderNode); !ok {
		return []domains.CachePlanCondition{}
	}

	return []domains.CachePlanCondition{
		{
			Column:      "OFFSET()",
			Operator:    domains.CachePlanOperator_EQ,
			Placeholder: domains.CachePlanPlaceholder{Index: index},
		},
	}
}

func (q *queryAnalyzer) analyzeEnum(order sql_parser.OrderEnum) (domains.CachePlanOrderEnum, error) {
	switch order {
	case sql_parser.Order_ASC:
		return domains.CachePlanOrder_ASC, nil
	case sql_parser.Order_DESC:
		return domains.CachePlanOrder_DESC, nil
	default:
		return "", fmt.Errorf("unknown order: %s", order)
	}
}

func (q *queryAnalyzer) analyzeOperator(node sql_parser.OperatorNode) (domains.CachePlanOperatorEnum, error) {
	switch node.Operator {
	case sql_parser.Operator_EQ:
		return domains.CachePlanOperator_EQ, nil
	case sql_parser.Operator_IN:
		return domains.CachePlanOperator_IN, nil
	default:
		return "", fmt.Errorf("unknown operator: %s", node.Operator)
	}
}
