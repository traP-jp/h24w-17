package sql_parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traP-jp/isuc/domains"
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
			node: UpdateStmtNode{Table: TableNode{Name: "users"}, Sets: UpdateSetsNode{Sets: []UpdateSetNode{
				{Column: ColumnNode{Name: "del_flg"}, Value: PlaceholderNode{}},
				{Column: ColumnNode{Name: "id"}, Value: PlaceholderNode{}},
			}}},
		},
		{
			query: "DELETE FROM livecomments WHERE id = ? AND livestream_id = ? AND (SELECT COUNT(*) FROM (SELECT ? AS text) AS texts INNER JOIN (SELECT CONCAT('%', ?, '%') AS pattern) AS patterns ON texts.text LIKE patterns.pattern) >= 1;",
			node:  DeleteStmtNode{Table: TableNode{Name: "livecomments"}},
		},
	}

	for _, test := range tests {
		node, err := ParseSQLWeekly(test.query, []domains.TableSchema{
			{TableName: "users", Columns: map[string]domains.TableSchemaColumn{
				"id":      {ColumnName: "id"},
				"del_flg": {ColumnName: "del_flg"},
			}},
		})
		assert.NoError(t, err)
		assert.Equal(t, test.node, node)
	}
}
