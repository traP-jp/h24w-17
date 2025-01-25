package normalizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeArgs(t *testing.T) {
	tests := []struct {
		query    string
		expected NormalizedArgs
	}{
		{
			query: "SELECT * FROM `table` WHERE `id` = ?;",
			expected: NormalizedArgs{
				Query:     "SELECT * FROM `table` WHERE `id` = ?;",
				ExtraArgs: []ExtraArg{},
			},
		},
		{
			query: "SELECT * FROM `table` WHERE `id` = 1;",
			expected: NormalizedArgs{
				Query:     "SELECT * FROM `table` WHERE `id` = ?;",
				ExtraArgs: []ExtraArg{{Column: "id", Value: 1}},
			},
		},
		{
			query: "SELECT * FROM `table` WHERE `id` = 1 AND `name` = 'foo';",
			expected: NormalizedArgs{
				Query:     "SELECT * FROM `table` WHERE `id` = ? AND `name` = ?;",
				ExtraArgs: []ExtraArg{{Column: "id", Value: 1}, {Column: "name", Value: "foo"}},
			},
		},
		{
			query: "SELECT * FROM `table` WHERE `id` = 1 AND `name` = ? AND `age` = 20;",
			expected: NormalizedArgs{
				Query:     "SELECT * FROM `table` WHERE `id` = ? AND `name` = ? AND `age` = ?;",
				ExtraArgs: []ExtraArg{{Column: "id", Value: 1}, {Column: "age", Value: 20}},
			},
		},
		{
			query: "SELECT * FROM `table` WHERE `id` IN (1, 2, 3);",
			expected: NormalizedArgs{
				Query: "SELECT * FROM `table` WHERE `id` IN (?, ?, ?);",
				ExtraArgs: []ExtraArg{
					{Column: "id", Value: 1},
					{Column: "id", Value: 2},
					{Column: "id", Value: 3},
				},
			},
		},
		{
			query: "SELECT * FROM `table` WHERE `id` IN (1, ?, 3);",
			expected: NormalizedArgs{
				Query: "SELECT * FROM `table` WHERE `id` IN (?, ?, ?);",
				ExtraArgs: []ExtraArg{
					{Column: "id", Value: 1},
					{Column: "id", Value: 3},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			actual, err := NormalizeArgs(test.query)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}
