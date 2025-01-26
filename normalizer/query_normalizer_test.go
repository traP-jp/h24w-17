package normalizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeQuery(t *testing.T) {
	tests := []struct {
		query    string
		expected string
	}{
		{
			query:    "   SELECT * from table;   ",
			expected: "SELECT * from table;",
		},
		{
			query:    "SELECT *   \t\n from table;",
			expected: "SELECT * from table;",
		},
		{
			query:    "SELECT `id` from table;",
			expected: "SELECT id from table;",
		},
		{
			query:    "INSERT INTO table (name, col) VALUES (?, ?);",
			expected: "INSERT INTO table (name, col) VALUES (?);",
		},
		{
			query:    "DELETE FROM table WHERE id IN (?, ?, ?, ?);",
			expected: "DELETE FROM table WHERE id IN (?);",
		},
		{
			query:    "SELECT * FROM table",
			expected: "SELECT * FROM table;",
		},
		{
			query:    "INSERT INTO users (name, display_name, description, password) VALUES(?, ?, ?, ?);",
			expected: "INSERT INTO users (name, display_name, description, password) VALUES(?);",
		},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			actual := NormalizeQuery(tt.query)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
