package sql_parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringer(t *testing.T) {
	tests := []struct {
		input    SQLNode
		expected string
	}{
		{
			input: SelectStmtNode{
				Values: SelectValuesNode{Values: []SQLNode{SelectValueColumnNode{Column: ColumnNode{Name: "id"}}}},
				Table:  TableNode{Name: "users"},
			},
			expected: "SELECT `id` FROM `users`;",
		},
		{
			expected: "SELECT * FROM `users` WHERE `id` = ?;",
			input: SelectStmtNode{
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
			input: SelectStmtNode{
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
			expected: "SELECT `name`, `age` FROM `users` WHERE `age` > 18 AND `name` LIKE '%test%';",
		},
		{
			input: SelectStmtNode{
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
			expected: "SELECT * FROM `users` WHERE `id` IN (1, 2, 3);",
		},
		{
			input: SelectStmtNode{
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
			expected: "SELECT COUNT(*) FROM `users` WHERE `id` = ? AND `name` = 'Alice' ORDER BY `id` ASC LIMIT 10 OFFSET 0;",
		},
		{
			input: UpdateStmtNode{
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
			expected: "UPDATE `users` SET `name` = 'Bob', `age` = 20 WHERE `id` = ?;",
		},
		{
			input: DeleteStmtNode{
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
			expected: "DELETE FROM `users` WHERE `id` = ?;",
		},
		{
			input: InsertStmtNode{
				Table: TableNode{Name: "users"},
				Columns: ColumnsNode{
					Columns: []ColumnNode{{Name: "name"}, {Name: "age"}},
				},
				Values: ValuesNode{
					Values: []SQLNode{StringNode{Value: "Cathy"}, NumberNode{Value: 30}},
				},
			},
			expected: "INSERT INTO `users` (`name`, `age`) VALUES ('Cathy', 30);",
		},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			assert.Equal(t, test.expected, test.input.String())
		})
	}
}
