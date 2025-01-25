package sql_parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	tests := []struct {
		name     string
		input    []token
		expected SQLNode
	}{
		{
			name: "SELECT id FROM users;",
			input: []token{
				{Type: tokenType_RESERVED, Literal: "SELECT"},
				{Type: tokenType_IDENTIFIER, Literal: "id"},
				{Type: tokenType_RESERVED, Literal: "FROM"},
				{Type: tokenType_IDENTIFIER, Literal: "users"},
				{Type: tokenType_SYMBOL, Literal: ";"},
				{Type: tokenType_EOF, Literal: ""},
			},
			expected: SelectStmtNode{
				Values: SelectValuesNode{Values: []SQLNode{SelectValueColumnNode{Column: ColumnNode{Name: "id"}}}},
				Table:  TableNode{Name: "users"},
			},
		},
		{
			name: "SELECT id AS id_alt, name name_alt FROM users;",
			input: []token{
				{Type: tokenType_RESERVED, Literal: "SELECT"},
				{Type: tokenType_IDENTIFIER, Literal: "id"},
				{Type: tokenType_RESERVED, Literal: "AS"},
				{Type: tokenType_IDENTIFIER, Literal: "id_alt"},
				{Type: tokenType_SYMBOL, Literal: ","},
				{Type: tokenType_IDENTIFIER, Literal: "name"},
				{Type: tokenType_IDENTIFIER, Literal: "name_alt"},
				{Type: tokenType_RESERVED, Literal: "FROM"},
				{Type: tokenType_IDENTIFIER, Literal: "users"},
				{Type: tokenType_SYMBOL, Literal: ";"},
				{Type: tokenType_EOF, Literal: ""},
			},
			expected: SelectStmtNode{
				Values: SelectValuesNode{
					Values: []SQLNode{
						SelectValueColumnNode{Column: ColumnNode{Name: "id"}},
						SelectValueColumnNode{Column: ColumnNode{Name: "name"}},
					},
				},
				Table: TableNode{Name: "users"},
			},
		},
		{
			name: "SELECT * FROM users WHERE id = ?",
			input: []token{
				{Type: tokenType_RESERVED, Literal: "SELECT"},
				{Type: tokenType_SYMBOL, Literal: "*"},
				{Type: tokenType_RESERVED, Literal: "FROM"},
				{Type: tokenType_IDENTIFIER, Literal: "users"},
				{Type: tokenType_RESERVED, Literal: "WHERE"},
				{Type: tokenType_IDENTIFIER, Literal: "id"},
				{Type: tokenType_SYMBOL, Literal: "="},
				{Type: tokenType_SYMBOL, Literal: "?"},
				{Type: tokenType_EOF, Literal: ""},
			},
			expected: SelectStmtNode{
				Values: SelectValuesNode{Values: []SQLNode{SelectValueAsteriskNode{}}},
				Table:  TableNode{Name: "users"},
				Conditions: &ConditionsNode{
					Conditions: []ConditionNode{
						{
							Column:   ColumnNode{Name: "id"},
							Operator: OperatorNode{Operator: Operator_EQ},
							Value:    PlaceholderNode{},
						},
					},
				},
			},
		},
		{
			name: "SELECT name, age FROM users WHERE age > 18 AND name LIKE '%test%';",
			input: []token{
				{Type: tokenType_RESERVED, Literal: "SELECT"},
				{Type: tokenType_IDENTIFIER, Literal: "name"},
				{Type: tokenType_SYMBOL, Literal: ","},
				{Type: tokenType_IDENTIFIER, Literal: "age"},
				{Type: tokenType_RESERVED, Literal: "FROM"},
				{Type: tokenType_IDENTIFIER, Literal: "users"},
				{Type: tokenType_RESERVED, Literal: "WHERE"},
				{Type: tokenType_IDENTIFIER, Literal: "age"},
				{Type: tokenType_SYMBOL, Literal: ">"},
				{Type: tokenType_NUMBER, Literal: "18"},
				{Type: tokenType_RESERVED, Literal: "AND"},
				{Type: tokenType_IDENTIFIER, Literal: "name"},
				{Type: tokenType_RESERVED, Literal: "LIKE"},
				{Type: tokenType_STRING, Literal: "%test%"},
				{Type: tokenType_SYMBOL, Literal: ";"},
				{Type: tokenType_EOF, Literal: ""},
			},
			expected: SelectStmtNode{
				Values: SelectValuesNode{Values: []SQLNode{
					SelectValueColumnNode{Column: ColumnNode{Name: "name"}},
					SelectValueColumnNode{Column: ColumnNode{Name: "age"}},
				}},
				Table: TableNode{Name: "users"},
				Conditions: &ConditionsNode{
					Conditions: []ConditionNode{
						{
							Column:   ColumnNode{Name: "age"},
							Operator: OperatorNode{Operator: Operator_GT},
							Value:    NumberNode{Value: 18},
						},
						{
							Column:   ColumnNode{Name: "name"},
							Operator: OperatorNode{Operator: Operator_LIKE},
							Value:    StringNode{Value: "%test%"},
						},
					},
				},
			},
		},
		{
			name: "SELECT * FROM users WHERE id IN (1, 2, 3);",
			input: []token{
				{Type: tokenType_RESERVED, Literal: "SELECT"},
				{Type: tokenType_SYMBOL, Literal: "*"},
				{Type: tokenType_RESERVED, Literal: "FROM"},
				{Type: tokenType_IDENTIFIER, Literal: "users"},
				{Type: tokenType_RESERVED, Literal: "WHERE"},
				{Type: tokenType_IDENTIFIER, Literal: "id"},
				{Type: tokenType_RESERVED, Literal: "IN"},
				{Type: tokenType_SYMBOL, Literal: "("},
				{Type: tokenType_NUMBER, Literal: "1"},
				{Type: tokenType_SYMBOL, Literal: ","},
				{Type: tokenType_NUMBER, Literal: "2"},
				{Type: tokenType_SYMBOL, Literal: ","},
				{Type: tokenType_NUMBER, Literal: "3"},
				{Type: tokenType_SYMBOL, Literal: ")"},
				{Type: tokenType_SYMBOL, Literal: ";"},
				{Type: tokenType_EOF, Literal: ""},
			},
			expected: SelectStmtNode{
				Values: SelectValuesNode{Values: []SQLNode{SelectValueAsteriskNode{}}},
				Table:  TableNode{Name: "users"},
				Conditions: &ConditionsNode{
					Conditions: []ConditionNode{
						{
							Column:   ColumnNode{Name: "id"},
							Operator: OperatorNode{Operator: Operator_IN},
							Value:    ValuesNode{Values: []SQLNode{NumberNode{Value: 1}, NumberNode{Value: 2}, NumberNode{Value: 3}}},
						},
					},
				},
			},
		},
		{
			name: "SELECT COUNT(*) FROM users WHERE id = ? AND name = 'Alice' ORDER BY id ASC LIMIT 10 OFFSET 0;",
			input: []token{
				{Type: tokenType_RESERVED, Literal: "SELECT"},
				{Type: tokenType_RESERVED, Literal: "COUNT"},
				{Type: tokenType_SYMBOL, Literal: "("},
				{Type: tokenType_SYMBOL, Literal: "*"},
				{Type: tokenType_SYMBOL, Literal: ")"},
				{Type: tokenType_RESERVED, Literal: "FROM"},
				{Type: tokenType_IDENTIFIER, Literal: "users"},
				{Type: tokenType_RESERVED, Literal: "WHERE"},
				{Type: tokenType_IDENTIFIER, Literal: "id"},
				{Type: tokenType_SYMBOL, Literal: "="},
				{Type: tokenType_SYMBOL, Literal: "?"},
				{Type: tokenType_RESERVED, Literal: "AND"},
				{Type: tokenType_IDENTIFIER, Literal: "name"},
				{Type: tokenType_SYMBOL, Literal: "="},
				{Type: tokenType_STRING, Literal: "Alice"},
				{Type: tokenType_RESERVED, Literal: "ORDER BY"},
				{Type: tokenType_IDENTIFIER, Literal: "id"},
				{Type: tokenType_RESERVED, Literal: "ASC"},
				{Type: tokenType_RESERVED, Literal: "LIMIT"},
				{Type: tokenType_NUMBER, Literal: "10"},
				{Type: tokenType_RESERVED, Literal: "OFFSET"},
				{Type: tokenType_NUMBER, Literal: "0"},
				{Type: tokenType_SYMBOL, Literal: ";"},
				{Type: tokenType_EOF, Literal: ""},
			},
			expected: SelectStmtNode{
				Values: SelectValuesNode{Values: []SQLNode{SelectValueFunctionNode{Name: "COUNT", Value: SelectValueAsteriskNode{}}}},
				Table:  TableNode{Name: "users"},
				Conditions: &ConditionsNode{
					Conditions: []ConditionNode{
						{
							Column:   ColumnNode{Name: "id"},
							Operator: OperatorNode{Operator: Operator_EQ},
							Value:    PlaceholderNode{},
						},
						{
							Column:   ColumnNode{Name: "name"},
							Operator: OperatorNode{Operator: Operator_EQ},
							Value:    StringNode{Value: "Alice"},
						},
					},
				},
				Orders: &OrdersNode{
					Orders: []OrderNode{
						{
							Column: ColumnNode{Name: "id"},
							Order:  Order_ASC,
						},
					},
				},
				Limit:  &LimitNode{Limit: NumberNode{Value: 10}},
				Offset: &OffsetNode{Offset: NumberNode{Value: 0}},
			},
		},
		{
			name: "UPDATE users SET name = 'Bob', age = 20 WHERE id = ?;",
			input: []token{
				{Type: tokenType_RESERVED, Literal: "UPDATE"},
				{Type: tokenType_IDENTIFIER, Literal: "users"},
				{Type: tokenType_RESERVED, Literal: "SET"},
				{Type: tokenType_IDENTIFIER, Literal: "name"},
				{Type: tokenType_SYMBOL, Literal: "="},
				{Type: tokenType_STRING, Literal: "Bob"},
				{Type: tokenType_SYMBOL, Literal: ","},
				{Type: tokenType_IDENTIFIER, Literal: "age"},
				{Type: tokenType_SYMBOL, Literal: "="},
				{Type: tokenType_NUMBER, Literal: "20"},
				{Type: tokenType_RESERVED, Literal: "WHERE"},
				{Type: tokenType_IDENTIFIER, Literal: "id"},
				{Type: tokenType_SYMBOL, Literal: "="},
				{Type: tokenType_SYMBOL, Literal: "?"},
				{Type: tokenType_SYMBOL, Literal: ";"},
				{Type: tokenType_EOF, Literal: ""},
			},
			expected: UpdateStmtNode{
				Table: TableNode{Name: "users"},
				Sets: UpdateSetsNode{
					Sets: []UpdateSetNode{
						{Column: ColumnNode{Name: "name"}, Value: StringNode{Value: "Bob"}},
						{Column: ColumnNode{Name: "age"}, Value: NumberNode{Value: 20}},
					},
				},
				Conditions: &ConditionsNode{
					Conditions: []ConditionNode{
						{
							Column:   ColumnNode{Name: "id"},
							Operator: OperatorNode{Operator: Operator_EQ},
							Value:    PlaceholderNode{},
						},
					},
				},
			},
		},
		{
			name: "DELETE FROM users WHERE id = ?;",
			input: []token{
				{Type: tokenType_RESERVED, Literal: "DELETE"},
				{Type: tokenType_RESERVED, Literal: "FROM"},
				{Type: tokenType_IDENTIFIER, Literal: "users"},
				{Type: tokenType_RESERVED, Literal: "WHERE"},
				{Type: tokenType_IDENTIFIER, Literal: "id"},
				{Type: tokenType_SYMBOL, Literal: "="},
				{Type: tokenType_SYMBOL, Literal: "?"},
				{Type: tokenType_SYMBOL, Literal: ";"},
				{Type: tokenType_EOF, Literal: ""},
			},
			expected: DeleteStmtNode{
				Table: TableNode{Name: "users"},
				Conditions: &ConditionsNode{
					Conditions: []ConditionNode{
						{
							Column:   ColumnNode{Name: "id"},
							Operator: OperatorNode{Operator: Operator_EQ},
							Value:    PlaceholderNode{},
						},
					},
				},
			},
		},
		{
			name: "INSERT INTO users (name, age) VALUES ('Cathy', 30);",
			input: []token{
				{Type: tokenType_RESERVED, Literal: "INSERT"},
				{Type: tokenType_RESERVED, Literal: "INTO"},
				{Type: tokenType_IDENTIFIER, Literal: "users"},
				{Type: tokenType_SYMBOL, Literal: "("},
				{Type: tokenType_IDENTIFIER, Literal: "name"},
				{Type: tokenType_SYMBOL, Literal: ","},
				{Type: tokenType_IDENTIFIER, Literal: "age"},
				{Type: tokenType_SYMBOL, Literal: ")"},
				{Type: tokenType_RESERVED, Literal: "VALUES"},
				{Type: tokenType_SYMBOL, Literal: "("},
				{Type: tokenType_STRING, Literal: "Cathy"},
				{Type: tokenType_SYMBOL, Literal: ","},
				{Type: tokenType_NUMBER, Literal: "30"},
				{Type: tokenType_SYMBOL, Literal: ")"},
				{Type: tokenType_SYMBOL, Literal: ";"},
				{Type: tokenType_EOF, Literal: ""},
			},
			expected: InsertStmtNode{
				Table: TableNode{Name: "users"},
				Columns: ColumnsNode{
					Columns: []ColumnNode{{Name: "name"}, {Name: "age"}},
				},
				Values: ValuesNode{
					Values: []SQLNode{StringNode{Value: "Cathy"}, NumberNode{Value: 30}},
				},
			},
		},
		{
			name: "SELECT COUNT(*) AS `count` FROM `comments` WHERE `post_id` = ?",
			input: []token{
				{Type: tokenType_RESERVED, Literal: "SELECT"},
				{Type: tokenType_RESERVED, Literal: "COUNT"},
				{Type: tokenType_SYMBOL, Literal: "("},
				{Type: tokenType_SYMBOL, Literal: "*"},
				{Type: tokenType_SYMBOL, Literal: ")"},
				{Type: tokenType_RESERVED, Literal: "AS"},
				{Type: tokenType_IDENTIFIER, Literal: "count"},
				{Type: tokenType_RESERVED, Literal: "FROM"},
				{Type: tokenType_IDENTIFIER, Literal: "comments"},
				{Type: tokenType_RESERVED, Literal: "WHERE"},
				{Type: tokenType_IDENTIFIER, Literal: "post_id"},
				{Type: tokenType_SYMBOL, Literal: "="},
				{Type: tokenType_SYMBOL, Literal: "?"},
				{Type: tokenType_EOF, Literal: ""},
			},
			expected: SelectStmtNode{
				Values: SelectValuesNode{Values: []SQLNode{SelectValueFunctionNode{Name: "COUNT", Value: SelectValueAsteriskNode{}}}},
				Table:  TableNode{Name: "comments"},
				Conditions: &ConditionsNode{
					Conditions: []ConditionNode{
						{Column: ColumnNode{Name: "post_id"}, Operator: OperatorNode{Operator: Operator_EQ}, Value: PlaceholderNode{}},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := NewParser(test.input)
			ast, err := p.Parse()
			assert.NoError(t, err)
			assert.Equal(t, test.expected, ast)
		})
	}
}
