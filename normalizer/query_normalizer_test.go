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
			query:    "SELECT `id` from `table`",
			expected: "SELECT `id` from `table`",
		},
		{
			query:    "INSERT INTO table (name, col) VALUES (?, ?)",
			expected: "INSERT INTO table (name, col) VALUES (?)",
		},
		{
			query:    "DELETE FROM `table` WHERE id IN (?, ?, ?, ?);",
			expected: "DELETE FROM `table` WHERE id IN (?);",
		},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			actual, err := NormalizeQuery(tt.query)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
