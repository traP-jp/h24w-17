package sql_parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSQLWeekly(t *testing.T) {
	tests := []struct {
		query string
		node  SQLNode
	}{
		{
			query: "SELECT r.emoji_name FROM users u INNER JOIN livestreams l ON l.user_id = u.id INNER JOIN reactions r ON r.livestream_id = l.id WHERE u.name = ? GROUP BY emoji_name ORDER BY COUNT(*) DESC, emoji_name DESC LIMIT ?",
			node:  SelectStmtNode{Table: TableNode{Name: "users"}},
		},
		{
			query: "INSERT INTO coupons (user_id, code, discount) VALUES (?, CONCAT(?, '_', FLOOR(UNIX_TIMESTAMP(NOW(3))*1000)), ?)",
			node:  InsertStmtNode{Table: TableNode{Name: "coupons"}},
		},
		{
			query: "UPDATE users SET del_flg = 1 WHERE id % 50 = 0",
			node:  UpdateStmtNode{Table: TableNode{Name: "users"}},
		},
	}

	for _, test := range tests {
		node, err := ParseSQLWeekly(test.query)
		assert.NoError(t, err)
		assert.Equal(t, test.node, node)
	}
}
