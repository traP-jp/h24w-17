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
			query:    "select id from table;",
			expected: "SELECT `id` FROM `table`;",
		},
		{
			query:    "insert into table (name,col) values (?, ?);",
			expected: "INSERT INTO `table` (`name`, `col`) VALUES (?);",
		},
		{
			query:    "update table set id = ? order by id desc limit 1 offset ?;",
			expected: "UPDATE `table` SET `id` = ? ORDER BY `id` DESC LIMIT 1 OFFSET ?;",
		},
		{
			query:    "delete from table where id in (?, ?, ?, ?);",
			expected: "DELETE FROM `table` WHERE `id` IN (?);",
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
