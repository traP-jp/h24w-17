package h24w17

import (
	"fmt"
	"testing"

	"github.com/h24w-17/domains"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeQuery(t *testing.T) {
	tests := []struct {
		query    string
		expected NormalizedQuery
	}{
		{
			query: "SELECT * FROM users WHERE id = ?;",
			expected: NormalizedQuery{
				Query:           "SELECT * FROM users WHERE id = ?;",
				ExtraConditions: []domains.CachePlanCondition{},
			},
		},
		{
			query: "SELECT * FROM users WHERE id = 1;",
			expected: NormalizedQuery{
				Query: "SELECT * FROM users WHERE id = ?;",
				ExtraConditions: []domains.CachePlanCondition{
					{Column: "id", Value: 1},
				},
			},
		},
		{
			query: "SELECT * FROM users WHERE id = 1 AND name = 'John';",
			expected: NormalizedQuery{
				Query: "SELECT * FROM users WHERE id = ? AND name = ?;",
				ExtraConditions: []domains.CachePlanCondition{
					{Column: "id", Value: 1},
					{Column: "name", Value: "John"},
				},
			},
		},
		{
			query: "SELECT * FROM users ORDER BY created_at LIMIT 1;",
			expected: NormalizedQuery{
				Query: "SELECT * FROM users ORDER BY created_at LIMIT ?;",
				ExtraConditions: []domains.CachePlanCondition{
					{Column: "LIMIT()", Value: 1},
				},
			},
		},
		{
			query: "SELECT * FROM users WHERE id IN (?, ?, ?);",
			expected: NormalizedQuery{
				Query:           "SELECT * FROM users WHERE id IN (?);",
				ExtraConditions: []domains.CachePlanCondition{},
			},
		},
		{
			query: "INSERT INTO users (id, name) VALUES (?, ?, ?, ?, ?, ?);",
			expected: NormalizedQuery{
				Query:           "INSERT INTO users (id, name) VALUES (?);",
				ExtraConditions: []domains.CachePlanCondition{},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("NormalizeQuery(\"%s\")", test.query), func(t *testing.T) {
			normalizedQuery, err := NormalizeQuery(test.query)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, normalizedQuery)
		})
	}
}
