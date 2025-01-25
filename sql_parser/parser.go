package sql_parser

import (
	"errors"
	"fmt"
	"strconv"
)

type SQLNode interface {
	String() string
}

type SelectStmtNode struct {
	Values     SelectValuesNode
	Table      TableNode
	Conditions *ConditionsNode
	Orders     *OrdersNode
	Limit      *LimitNode
	Offset     *OffsetNode
}

type SelectValuesNode struct {
	// SelectValueAsteriskNode | SelectValueColumnNode | SelectValueFunctionNode | StringNode | NumberNode
	Values []SQLNode
}

type SelectValueAsteriskNode struct{}

type SelectValueColumnNode struct {
	Column ColumnNode
}

type SelectValueFunctionNode struct {
	Name string
	// SelectValueAsteriskNode | SelectValueColumnNode | SelectValueFunctionNode
	Value SQLNode
}

type UpdateStmtNode struct {
	Table      TableNode
	Sets       UpdateSetsNode
	Conditions *ConditionsNode
	Orders     *OrdersNode
	Limit      *LimitNode
	Offset     *OffsetNode
}

type UpdateSetsNode struct {
	Sets []UpdateSetNode
}

type UpdateSetNode struct {
	Column ColumnNode
	// StringNode | NumberNode | PlaceholderNode
	Value SQLNode
}

type DeleteStmtNode struct {
	Table      TableNode
	Conditions *ConditionsNode
	Orders     *OrdersNode
	Limit      *LimitNode
	Offset     *OffsetNode
}

type InsertStmtNode struct {
	Table   TableNode
	Columns ColumnsNode
	Values  ValuesNode
}

type ConditionsNode struct {
	Conditions []ConditionNode
}

type ConditionNode struct {
	Column   ColumnNode
	Operator OperatorNode
	// StringNode | NumberNode | PlaceholderNode | ValuesNode
	Value SQLNode
}

type OrdersNode struct {
	Orders []OrderNode
}

type OrderNode struct {
	Column ColumnNode
	Order  OrderEnum
}

type OrderEnum string

const (
	Order_ASC  OrderEnum = "ASC"
	Order_DESC OrderEnum = "DESC"
)

type LimitNode struct {
	// NumberNode | PlaceholderNode
	Limit SQLNode
}

type OffsetNode struct {
	// NumberNode | PlaceholderNode
	Offset SQLNode
}

type OperatorNode struct {
	Operator OperatorEnum
}

type OperatorEnum string

const (
	Operator_EQ   OperatorEnum = "="
	Operator_NEQ  OperatorEnum = "!="
	Operator_LT   OperatorEnum = "<"
	Operator_GT   OperatorEnum = ">"
	Operator_LTE  OperatorEnum = "<="
	Operator_GTE  OperatorEnum = ">="
	Operator_LIKE OperatorEnum = "LIKE"
	Operator_IN   OperatorEnum = "IN"
)

type ColumnsNode struct {
	Columns []ColumnNode
}

type ColumnNode struct {
	Name string
}

type TableNode struct {
	Name string
}

type ValuesNode struct {
	// stringNode | numberNode | placeholderNode
	Values []SQLNode
}

type StringNode struct {
	Value string
}

type NumberNode struct {
	Value int
}

type PlaceholderNode struct{}

type parser struct {
	tokens []token
	cursor int
}

func NewParser(tokens []token) *parser {
	return &parser{tokens: tokens, cursor: 0}
}

func (p *parser) Parse() (ast SQLNode, err error) {
	defer func() {
		if r := recover(); r != nil {
			ast = nil
			err = fmt.Errorf("unexpected error -> %v", r)
		}
	}()
	ast, err = p.sql()
	if err != nil {
		return nil, err
	}
	if !p.expect(token{Type: tokenType_EOF, Literal: ""}) {
		return nil, fmt.Errorf("no more tokens expected, got %v", p.peek().String())
	}
	return ast, nil
}

func (p *parser) peek() token {
	if len(p.tokens) == p.cursor {
		panic("peek() -> unexpected EOF")
	}
	return p.tokens[p.cursor]
}

func (p *parser) check(expected token) bool {
	if len(p.tokens) == p.cursor {
		panic("check() -> unexpected EOF")
	}
	actual := p.tokens[p.cursor]
	return expected.Type == actual.Type && expected.Literal == actual.Literal
}

func (p *parser) expect(expected token) bool {
	ok := p.check(expected)
	if ok {
		p.cursor++
	}
	return ok
}

func (p *parser) consume() token {
	if len(p.tokens) == p.cursor {
		panic("consume() -> unexpected EOF")
	}
	t := p.tokens[p.cursor]
	p.cursor++
	return t
}

func (p *parser) sql() (SQLNode, error) {
	t := p.peek()
	if t.Type != tokenType_RESERVED {
		return nil, errors.New("<sql> is not valid")
	}

	switch t.Literal {
	case "SELECT":
		return p.selectStmt()
	case "UPDATE":
		return p.updateStmt()
	case "DELETE":
		return p.deleteStmt()
	case "INSERT":
		return p.insertStmt()
	default:
		return nil, fmt.Errorf("<sql> got unexpected token %v", t.String())
	}
}

func (p *parser) selectStmt() (SelectStmtNode, error) {
	node := SelectStmtNode{}

	if !p.expect(token{Type: tokenType_RESERVED, Literal: "SELECT"}) {
		return SelectStmtNode{}, fmt.Errorf("<select-stmt> expected <reserved(SELECT)>, got %v", p.peek().String())
	}

	values, err := p.selectValues()
	if err != nil {
		return SelectStmtNode{}, fmt.Errorf("<select-stmt> %v", err)
	}
	node.Values = values

	if !p.expect(token{Type: tokenType_RESERVED, Literal: "FROM"}) {
		return SelectStmtNode{}, fmt.Errorf("<select-stmt> expected <reserved(FROM)>, got %v", p.peek().String())
	}

	table, err := p.table()
	if err != nil {
		return SelectStmtNode{}, fmt.Errorf("<select-stmt> %v", err)
	}
	node.Table = table

	if p.expect(token{Type: tokenType_RESERVED, Literal: "WHERE"}) {
		conditions, err := p.conditions()
		if err != nil {
			return SelectStmtNode{}, fmt.Errorf("<select-stmt> %v", err)
		}
		node.Conditions = &conditions
	}

	if p.expect(token{Type: tokenType_RESERVED, Literal: "ORDER BY"}) {
		orders, err := p.orders()
		if err != nil {
			return SelectStmtNode{}, fmt.Errorf("<select-stmt> %v", err)
		}
		node.Orders = &orders
	}

	if p.expect(token{Type: tokenType_RESERVED, Literal: "LIMIT"}) {
		limit, err := p.limit()
		if err != nil {
			return SelectStmtNode{}, fmt.Errorf("<select-stmt> %v", err)
		}
		node.Limit = &limit
	}

	if p.expect(token{Type: tokenType_RESERVED, Literal: "OFFSET"}) {
		offset, err := p.offset()
		if err != nil {
			return SelectStmtNode{}, fmt.Errorf("<select-stmt> %v", err)
		}
		node.Offset = &offset
	}

	p.expect(token{Type: tokenType_SYMBOL, Literal: ";"})
	return node, nil
}

func (p *parser) selectValues() (SelectValuesNode, error) {
	node := SelectValuesNode{}
	selectValueNode, err := p.selectValue()
	if err != nil {
		return SelectValuesNode{}, fmt.Errorf("<select-values> %v", err)
	}
	node.Values = append(node.Values, selectValueNode)
	if p.expect(token{Type: tokenType_SYMBOL, Literal: ","}) {
		nextNode, err := p.selectValues()
		if err != nil {
			return SelectValuesNode{}, fmt.Errorf("<select-values> %v", err)
		}
		node.Values = append(node.Values, nextNode.Values...)
		return node, nil
	}
	return node, nil
}

func (p *parser) selectValue() (SQLNode, error) {
	if p.expect(token{Type: tokenType_SYMBOL, Literal: "*"}) {
		p.selectAlias()
		return SelectValueAsteriskNode{}, nil
	}
	t := p.peek()
	if t.Type == tokenType_RESERVED {
		switch t.Literal {
		case "COUNT", "SUM", "AVG", "MIN", "MAX":
			t = p.consume()
			if !p.expect(token{Type: tokenType_SYMBOL, Literal: "("}) {
				return nil, fmt.Errorf("<select-value> expected <symbol(()>, got %v", t.String())
			}
			v, err := p.selectValue()
			if err != nil {
				return nil, fmt.Errorf("<select-value> %v", err)
			}
			if !p.expect(token{Type: tokenType_SYMBOL, Literal: ")"}) {
				return nil, fmt.Errorf("<select-value> expected <symbol())>, got %v", t.String())
			}
			p.selectAlias()
			return SelectValueFunctionNode{Name: t.Literal, Value: v}, nil
		}
	}
	if t.Type == tokenType_STRING {
		p.consume()
		return StringNode{Value: t.Literal}, nil
	}
	if t.Type == tokenType_NUMBER {
		p.consume()
		parsed, err := strconv.Atoi(t.Literal)
		if err != nil {
			return nil, fmt.Errorf("<select-value> failed to parse number %v", err)
		}
		return NumberNode{Value: parsed}, nil
	}

	column, err := p.column()
	if err == nil {
		p.selectAlias()
		return SelectValueColumnNode{Column: column}, nil
	}
	return nil, fmt.Errorf("<select-value> got unexpected token %v", t.String())
}

func (p *parser) selectAlias() (SQLNode, error) {
	// alias is not used
	if p.expect(token{Type: tokenType_RESERVED, Literal: "AS"}) {
		_, err := p.column()
		if err != nil {
			return nil, fmt.Errorf("<select-alias> %v", err)
		}
		return nil, nil
	} else {
		p.column()
	}
	return nil, nil
}

func (p *parser) updateStmt() (UpdateStmtNode, error) {
	node := UpdateStmtNode{}
	if !p.expect(token{Type: tokenType_RESERVED, Literal: "UPDATE"}) {
		return UpdateStmtNode{}, fmt.Errorf("<update-stmt> expected <reserved(UPDATE)>, got %v", p.peek().String())
	}

	table, err := p.table()
	if err != nil {
		return UpdateStmtNode{}, fmt.Errorf("<update-stmt> %v", err)
	}
	node.Table = table

	if !p.expect(token{Type: tokenType_RESERVED, Literal: "SET"}) {
		return UpdateStmtNode{}, fmt.Errorf("<update-stmt> expected <reserved(SET)>, got %v", p.peek().String())
	}

	sets, err := p.updateSets()
	if err != nil {
		return UpdateStmtNode{}, fmt.Errorf("<update-stmt> %v", err)
	}
	node.Sets = sets

	if p.expect(token{Type: tokenType_RESERVED, Literal: "WHERE"}) {
		conditions, err := p.conditions()
		if err != nil {
			return UpdateStmtNode{}, fmt.Errorf("<update-stmt> %v", err)
		}
		node.Conditions = &conditions
	}

	if p.expect(token{Type: tokenType_RESERVED, Literal: "ORDER BY"}) {
		orders, err := p.orders()
		if err != nil {
			return UpdateStmtNode{}, fmt.Errorf("<update-stmt> %v", err)
		}
		node.Orders = &orders
	}

	if p.expect(token{Type: tokenType_RESERVED, Literal: "LIMIT"}) {
		limit, err := p.limit()
		if err != nil {
			return UpdateStmtNode{}, fmt.Errorf("<update-stmt> %v", err)
		}
		node.Limit = &limit
	}

	if p.expect(token{Type: tokenType_RESERVED, Literal: "OFFSET"}) {
		offset, err := p.offset()
		if err != nil {
			return UpdateStmtNode{}, fmt.Errorf("<update-stmt> %v", err)
		}
		node.Offset = &offset
	}

	p.expect(token{Type: tokenType_SYMBOL, Literal: ";"})
	return node, nil
}

func (p *parser) updateSets() (UpdateSetsNode, error) {
	node := UpdateSetsNode{}
	updateSet, err := p.updateSet()
	if err != nil {
		return UpdateSetsNode{}, fmt.Errorf("<update-sets> %v", err)
	}
	node.Sets = append(node.Sets, updateSet)

	if p.expect(token{Type: tokenType_SYMBOL, Literal: ","}) {
		nextNode, err := p.updateSets()
		if err != nil {
			return UpdateSetsNode{}, fmt.Errorf("<update-sets> %v", err)
		}
		node.Sets = append(node.Sets, nextNode.Sets...)
	}
	return node, nil
}

func (p *parser) updateSet() (UpdateSetNode, error) {
	column, err := p.column()
	if err != nil {
		return UpdateSetNode{}, fmt.Errorf("<update-set> %v", err)
	}

	if !p.expect(token{Type: tokenType_SYMBOL, Literal: "="}) {
		return UpdateSetNode{}, fmt.Errorf("<update-set> expected <symbol(=)>, got %v", p.peek().String())
	}

	value, err := p.value()
	if err != nil {
		return UpdateSetNode{}, fmt.Errorf("<update-set> %v", err)
	}

	return UpdateSetNode{Column: column, Value: value}, nil
}

func (p *parser) deleteStmt() (SQLNode, error) {
	node := DeleteStmtNode{}
	if !p.expect(token{Type: tokenType_RESERVED, Literal: "DELETE"}) {
		return nil, fmt.Errorf("<delete-stmt> expected <reserved(DELETE)>, got %v", p.peek().String())
	}

	if !p.expect(token{Type: tokenType_RESERVED, Literal: "FROM"}) {
		return nil, fmt.Errorf("<delete-stmt> expected <reserved(FROM)>, got %v", p.peek().String())
	}

	table, err := p.table()
	if err != nil {
		return nil, fmt.Errorf("<delete-stmt> %v", err)
	}
	node.Table = table

	if p.expect(token{Type: tokenType_RESERVED, Literal: "WHERE"}) {
		conditions, err := p.conditions()
		if err != nil {
			return nil, fmt.Errorf("<delete-stmt> %v", err)
		}
		node.Conditions = &conditions
	}

	if p.expect(token{Type: tokenType_RESERVED, Literal: "ORDER BY"}) {
		orders, err := p.orders()
		if err != nil {
			return nil, fmt.Errorf("<delete-stmt> %v", err)
		}
		node.Orders = &orders
	}

	if p.expect(token{Type: tokenType_RESERVED, Literal: "LIMIT"}) {
		limit, err := p.limit()
		if err != nil {
			return nil, fmt.Errorf("<delete-stmt> %v", err)
		}
		node.Limit = &limit
	}

	if p.expect(token{Type: tokenType_RESERVED, Literal: "OFFSET"}) {
		offset, err := p.offset()
		if err != nil {
			return nil, fmt.Errorf("<delete-stmt> %v", err)
		}
		node.Offset = &offset
	}

	p.expect(token{Type: tokenType_SYMBOL, Literal: ";"})
	return node, nil
}

func (p *parser) insertStmt() (SQLNode, error) {
	node := InsertStmtNode{}
	if !p.expect(token{Type: tokenType_RESERVED, Literal: "INSERT"}) {
		return nil, fmt.Errorf("<insert-stmt> expected <reserved(INSERT)>, got %v", p.peek().String())
	}

	if !p.expect(token{Type: tokenType_RESERVED, Literal: "INTO"}) {
		return nil, fmt.Errorf("<insert-stmt> expected <reserved(INTO)>, got %v", p.peek().String())
	}

	table, err := p.table()
	if err != nil {
		return nil, fmt.Errorf("<insert-stmt> %v", err)
	}
	node.Table = table

	if !p.expect(token{Type: tokenType_SYMBOL, Literal: "("}) {
		return nil, fmt.Errorf("<insert-stmt> expected <symbol(()>, got %v", p.peek().String())
	}

	columns, err := p.columns()
	if err != nil {
		return nil, fmt.Errorf("<insert-stmt> %v", err)
	}
	node.Columns = columns

	if !p.expect(token{Type: tokenType_SYMBOL, Literal: ")"}) {
		return nil, fmt.Errorf("<insert-stmt> expected <symbol())>, got %v", p.peek().String())
	}

	if !p.expect(token{Type: tokenType_RESERVED, Literal: "VALUES"}) {
		return nil, fmt.Errorf("<insert-stmt> expected <reserved(VALUES)>, got %v", p.peek().String())
	}

	if !p.expect(token{Type: tokenType_SYMBOL, Literal: "("}) {
		return nil, fmt.Errorf("<insert-stmt> expected <symbol(()>, got %v", p.peek().String())
	}

	values, err := p.values()
	if err != nil {
		return nil, fmt.Errorf("<insert-stmt> %v", err)
	}
	node.Values = values

	if !p.expect(token{Type: tokenType_SYMBOL, Literal: ")"}) {
		return nil, fmt.Errorf("<insert-stmt> expected <symbol())>, got %v", p.peek().String())
	}

	p.expect(token{Type: tokenType_SYMBOL, Literal: ";"})
	return node, nil
}

func (p *parser) conditions() (ConditionsNode, error) {
	node := ConditionsNode{}
	condition, err := p.condition()
	if err != nil {
		return ConditionsNode{}, fmt.Errorf("<conditions> %v", err)
	}
	node.Conditions = append(node.Conditions, condition)

	if p.expect(token{Type: tokenType_RESERVED, Literal: "AND"}) {
		nextNode, err := p.conditions()
		if err != nil {
			return ConditionsNode{}, fmt.Errorf("<conditions> %v", err)
		}
		node.Conditions = append(node.Conditions, nextNode.Conditions...)
	}
	return node, nil
}

func (p *parser) condition() (ConditionNode, error) {
	column, err := p.column()
	if err != nil {
		return ConditionNode{}, fmt.Errorf("<condition> %v", err)
	}

	t := p.consume()
	operators := []struct {
		token    token
		operator OperatorEnum
	}{
		{token{Type: tokenType_SYMBOL, Literal: "="}, Operator_EQ},
		{token{Type: tokenType_SYMBOL, Literal: "!="}, Operator_NEQ},
		{token{Type: tokenType_SYMBOL, Literal: "<"}, Operator_LT},
		{token{Type: tokenType_SYMBOL, Literal: ">"}, Operator_GT},
		{token{Type: tokenType_SYMBOL, Literal: "<="}, Operator_LTE},
		{token{Type: tokenType_SYMBOL, Literal: ">="}, Operator_GTE},
		{token{Type: tokenType_RESERVED, Literal: "LIKE"}, Operator_LIKE},
		{token{Type: tokenType_RESERVED, Literal: "IN"}, Operator_IN},
	}
	var operator OperatorEnum
	for _, op := range operators {
		if t.Type == op.token.Type && t.Literal == op.token.Literal {
			operator = op.operator
			break
		}
	}
	if operator == "" {
		return ConditionNode{}, fmt.Errorf("<condition> expected operators, got unexpected token %v", t.String())
	}

	value, err := p.value()
	if err != nil {
		return ConditionNode{}, fmt.Errorf("<condition> %v", err)
	}

	return ConditionNode{Column: column, Operator: OperatorNode{Operator: operator}, Value: value}, nil
}

func (p *parser) orders() (OrdersNode, error) {
	node := OrdersNode{}
	order, err := p.order()
	if err != nil {
		return OrdersNode{}, fmt.Errorf("<orders> %v", err)
	}
	node.Orders = append(node.Orders, order)
	if p.expect(token{Type: tokenType_SYMBOL, Literal: ","}) {
		nextNode, err := p.orders()
		if err != nil {
			return OrdersNode{}, fmt.Errorf("<orders> %v", err)
		}
		node.Orders = append(node.Orders, nextNode.Orders...)
		return node, nil
	}
	return node, nil
}

func (p *parser) order() (OrderNode, error) {
	column, err := p.column()
	if err != nil {
		return OrderNode{}, fmt.Errorf("<order> %v", err)
	}

	t := p.consume()
	if t.Type == tokenType_RESERVED && (t.Literal == "ASC" || t.Literal == "DESC") {
		return OrderNode{Column: column, Order: OrderEnum(t.Literal)}, nil
	}
	return OrderNode{Column: column, Order: Order_ASC}, nil
}

func (p *parser) limit() (LimitNode, error) {
	t := p.peek()
	if t.Type == tokenType_NUMBER {
		p.consume()
		parsed, err := strconv.Atoi(t.Literal)
		if err != nil {
			return LimitNode{}, fmt.Errorf("<limit> failed to parse number %v", err)
		}
		return LimitNode{Limit: NumberNode{Value: parsed}}, nil
	}
	if p.expect(token{Type: tokenType_SYMBOL, Literal: "?"}) {
		return LimitNode{Limit: PlaceholderNode{}}, nil
	}
	return LimitNode{}, fmt.Errorf("<limit> got unexpected token %v", t.String())
}

func (p *parser) offset() (OffsetNode, error) {
	t := p.peek()
	if t.Type == tokenType_NUMBER {
		p.consume()
		parsed, err := strconv.Atoi(t.Literal)
		if err != nil {
			return OffsetNode{}, fmt.Errorf("<offset> failed to parse number %v", err)
		}
		return OffsetNode{Offset: NumberNode{Value: parsed}}, nil
	}
	if p.expect(token{Type: tokenType_SYMBOL, Literal: "?"}) {
		return OffsetNode{Offset: PlaceholderNode{}}, nil
	}
	return OffsetNode{}, fmt.Errorf("<offset> got unexpected token %v", t.String())
}

func (p *parser) columns() (ColumnsNode, error) {
	node := ColumnsNode{}
	column, err := p.column()
	if err != nil {
		return ColumnsNode{}, fmt.Errorf("<columns> %v", err)
	}
	node.Columns = append(node.Columns, column)
	if p.expect(token{Type: tokenType_SYMBOL, Literal: ","}) {
		nextNode, err := p.columns()
		if err != nil {
			return ColumnsNode{}, fmt.Errorf("<columns> %v", err)
		}
		node.Columns = append(node.Columns, nextNode.Columns...)
		return node, nil
	}
	return node, nil
}

func (p *parser) column() (ColumnNode, error) {
	t := p.peek()
	if t.Type == tokenType_IDENTIFIER {
		p.consume()
		return ColumnNode{Name: t.Literal}, nil
	}
	if p.expect(token{Type: tokenType_SYMBOL, Literal: "`"}) {
		t := p.consume()
		if t.Type != tokenType_IDENTIFIER {
			return ColumnNode{}, fmt.Errorf("<column> expected <identifier>, got %v", t.String())
		}
		column := ColumnNode{Name: t.Literal}
		if !p.expect(token{Type: tokenType_SYMBOL, Literal: "`"}) {
			return ColumnNode{}, fmt.Errorf("<column> expected <symbol(`)>, got %v", t.String())
		}
		return column, nil
	}
	return ColumnNode{}, fmt.Errorf("<column> got unexpected token %v", t.String())
}

func (p *parser) table() (TableNode, error) {
	t := p.peek()
	if t.Type == tokenType_IDENTIFIER {
		p.consume()
		return TableNode{Name: t.Literal}, nil
	}
	if p.expect(token{Type: tokenType_SYMBOL, Literal: "`"}) {
		t := p.consume()
		if t.Type != tokenType_IDENTIFIER {
			return TableNode{}, fmt.Errorf("<table> expected <identifier>, got %v", t.String())
		}
		table := TableNode{Name: t.Literal}
		if !p.expect(token{Type: tokenType_SYMBOL, Literal: "`"}) {
			return TableNode{}, fmt.Errorf("<table> expected <symbol(`)>, got %v", t.String())
		}
		return table, nil
	}
	return TableNode{}, fmt.Errorf("<table> got unexpected token %v", t.String())
}

func (p *parser) values() (ValuesNode, error) {
	node := ValuesNode{}
	value, err := p.value()
	if err != nil {
		return ValuesNode{}, fmt.Errorf("<values> %v", err)
	}
	node.Values = append(node.Values, value)
	if p.expect(token{Type: tokenType_SYMBOL, Literal: ","}) {
		nextNode, err := p.values()
		if err != nil {
			return ValuesNode{}, fmt.Errorf("<values> %v", err)
		}
		node.Values = append(node.Values, nextNode.Values...)
		return node, nil
	}
	return node, nil
}

func (p *parser) value() (SQLNode, error) {
	t := p.peek()
	switch t.Type {
	case tokenType_STRING:
		p.consume()
		return StringNode{Value: t.Literal}, nil
	case tokenType_NUMBER:
		p.consume()
		parsed, err := strconv.Atoi(t.Literal)
		if err != nil {
			return nil, fmt.Errorf("<value> failed to parse number %v", err)
		}
		return NumberNode{Value: parsed}, nil
	case tokenType_SYMBOL:
		if t.Literal == "?" {
			p.consume()
			return PlaceholderNode{}, nil
		}
		if t.Literal == "(" {
			p.consume()
			values, err := p.values()
			if err != nil {
				return nil, fmt.Errorf("<value> %v", err)
			}
			if !p.expect(token{Type: tokenType_SYMBOL, Literal: ")"}) {
				return nil, fmt.Errorf("<value> expected <symbol())>, got %v", t.String())
			}
			return values, nil
		}
	}
	return nil, fmt.Errorf("<value> got unexpected token %v", t.String())
}

func (p *parser) diagnostics() string {
	diagnostics := ""
	start := 0
	end := 0
	for i, t := range p.tokens {
		if i == p.cursor {
			start = len(diagnostics)
			end = start + len(t.toSQL())
		}
		diagnostics += t.toSQL() + " "
	}
	diagnostics += "\n"
	for i := 0; i < start; i++ {
		diagnostics += " "
	}
	for i := start; i < end; i++ {
		diagnostics += "^"
	}
	diagnostics += "\n"
	return diagnostics
}
