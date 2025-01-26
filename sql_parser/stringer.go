package sql_parser

import "fmt"

var _ SQLNode = SelectStmtNode{}

func (n SelectStmtNode) String() string {
	sql := fmt.Sprintf("SELECT %s FROM %s", n.Values.String(), n.Table.String())
	if n.Conditions != nil {
		sql += fmt.Sprintf(" WHERE %s", n.Conditions.String())
	}
	if n.Orders != nil {
		sql += fmt.Sprintf(" ORDER BY %s", n.Orders.String())
	}
	if n.Limit != nil {
		sql += fmt.Sprintf(" LIMIT %s", n.Limit.String())
	}
	if n.Offset != nil {
		sql += fmt.Sprintf(" OFFSET %s", n.Offset.String())
	}
	sql += ";"
	return sql
}

var _ SQLNode = SelectValuesNode{}

func (n SelectValuesNode) String() string {
	sql := ""
	for i, v := range n.Values {
		if i > 0 {
			sql += ", "
		}
		sql += v.String()
	}
	return sql
}

var _ SQLNode = SelectValueAsteriskNode{}

func (n SelectValueAsteriskNode) String() string {
	return "*"
}

var _ SQLNode = SelectValueColumnNode{}

func (n SelectValueColumnNode) String() string {
	return n.Column.String()
}

var _ SQLNode = SelectValueFunctionNode{}

func (n SelectValueFunctionNode) String() string {
	return fmt.Sprintf("%s(%s)", n.Name, n.Value.String())
}

var _ SQLNode = UpdateStmtNode{}

func (n UpdateStmtNode) String() string {
	sql := fmt.Sprintf("UPDATE %s SET %s", n.Table.String(), n.Sets.String())
	if n.Conditions != nil {
		sql += fmt.Sprintf(" WHERE %s", n.Conditions.String())
	}
	if n.Orders != nil {
		sql += fmt.Sprintf(" ORDER BY %s", n.Orders.String())
	}
	if n.Limit != nil {
		sql += fmt.Sprintf(" LIMIT %s", n.Limit.String())
	}
	if n.Offset != nil {
		sql += fmt.Sprintf(" OFFSET %s", n.Offset.String())
	}
	sql += ";"
	return sql
}

var _ SQLNode = UpdateSetsNode{}

func (n UpdateSetsNode) String() string {
	sql := ""
	for i, s := range n.Sets {
		if i > 0 {
			sql += ", "
		}
		sql += s.String()
	}
	return sql
}

var _ SQLNode = UpdateSetNode{}

func (n UpdateSetNode) String() string {
	return fmt.Sprintf("%s = %s", n.Column.String(), n.Value.String())
}

var _ SQLNode = DeleteStmtNode{}

func (n DeleteStmtNode) String() string {
	sql := fmt.Sprintf("DELETE FROM %s", n.Table.String())
	if n.Conditions != nil {
		sql += fmt.Sprintf(" WHERE %s", n.Conditions.String())
	}
	if n.Orders != nil {
		sql += fmt.Sprintf(" ORDER BY %s", n.Orders.String())
	}
	if n.Limit != nil {
		sql += fmt.Sprintf(" LIMIT %s", n.Limit.String())
	}
	if n.Offset != nil {
		sql += fmt.Sprintf(" OFFSET %s", n.Offset.String())
	}
	sql += ";"
	return sql
}

var _ SQLNode = InsertStmtNode{}

func (n InsertStmtNode) String() string {
	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s;", n.Table.String(), n.Columns.String(), n.Values.String())
	return sql
}

var _ SQLNode = ConditionsNode{}

func (n ConditionsNode) String() string {
	sql := ""
	for i, c := range n.Conditions {
		if i > 0 {
			sql += " AND "
		}
		sql += c.String()
	}
	return sql
}

var _ SQLNode = ConditionNode{}

func (n ConditionNode) String() string {
	return fmt.Sprintf("%s %s %s", n.Column.String(), n.Operator.String(), n.Value.String())
}

var _ SQLNode = OrdersNode{}

func (n OrdersNode) String() string {
	sql := ""
	for i, o := range n.Orders {
		if i > 0 {
			sql += ", "
		}
		sql += o.String()
	}
	return sql
}

var _ SQLNode = OrderNode{}

func (n OrderNode) String() string {
	sql := n.Column.String()
	if n.Order != "" {
		sql += " " + string(n.Order)
	}
	return sql
}

var _ SQLNode = LimitNode{}

func (n LimitNode) String() string {
	return n.Limit.String()
}

var _ SQLNode = OffsetNode{}

func (n OffsetNode) String() string {
	return n.Offset.String()
}

var _ SQLNode = OperatorNode{}

func (n OperatorNode) String() string {
	return string(n.Operator)
}

var _ SQLNode = ColumnsNode{}

func (n ColumnsNode) String() string {
	sql := ""
	for i, c := range n.Columns {
		if i > 0 {
			sql += ", "
		}
		sql += c.String()
	}
	return sql
}

var _ SQLNode = ColumnNode{}

func (n ColumnNode) String() string {
	return n.Name
}

var _ SQLNode = TableNode{}

func (n TableNode) String() string {
	return n.Name
}

var _ SQLNode = ValuesNode{}

func (n ValuesNode) String() string {
	sql := "("
	for i, v := range n.Values {
		if i > 0 {
			sql += ", "
		}
		sql += v.String()
	}
	sql += ")"
	return sql
}

var _ SQLNode = StringNode{}

func (n StringNode) String() string {
	return fmt.Sprintf("'%s'", n.Value)
}

var _ SQLNode = NumberNode{}

func (n NumberNode) String() string {
	return fmt.Sprintf("%d", n.Value)
}

var _ SQLNode = PlaceholderNode{}

func (n PlaceholderNode) String() string {
	return "?"
}
