package analyzer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traP-jp/isuc/domains"
)

func TestAnalyzeQueries(t *testing.T) {
	tests := []struct {
		name     string
		queries  []string
		schemas  []domains.TableSchema
		expected domains.CachePlan
	}{
		{
			name: "simple",
			queries: []string{
				"SELECT `id` FROM `users`",
				"SELECT * FROM `users` WHERE `id` = ?",
				"SELECT * FROM `users` WHERE `id` = ? AND `name` = ?",
				"SELECT * FROM `users` ORDER BY created_at DESC LIMIT 10 OFFSET ?",
				"UPDATE `users` SET `name` = ? WHERE `id` = ?",
				"UPDATE `users` SET `name` = 'Alice' WHERE `id` IN (?)",
				"DELETE FROM `users` WHERE `id` = ?",
				"DELETE FROM `users` WHERE `id` IN (?)",
				"INSERT INTO `users` (`name`, `username`, `created_at`) VALUES (?)",
			},
			schemas: []domains.TableSchema{
				{
					TableName: "users",
					Columns: map[string]domains.TableSchemaColumn{
						"id":         {ColumnName: "id", DataType: domains.TableSchemaDataType_INT, IsNullable: false, IsPrimary: true, IsUnique: false},
						"name":       {ColumnName: "name", DataType: domains.TableSchemaDataType_STRING, IsNullable: false, IsPrimary: false, IsUnique: true},
						"username":   {ColumnName: "username", DataType: domains.TableSchemaDataType_STRING, IsNullable: false, IsPrimary: false, IsUnique: false},
						"created_at": {ColumnName: "created_at", DataType: domains.TableSchemaDataType_DATETIME, IsNullable: false, IsPrimary: false, IsUnique: false},
					},
				},
			},
			expected: domains.CachePlan{
				Queries: []*domains.CachePlanQuery{
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "SELECT id FROM users;",
							Type:  domains.CachePlanQueryType_SELECT,
						},
						Select: &domains.CachePlanSelectQuery{
							Cache:      true,
							Table:      "users",
							Targets:    []string{"id"},
							Conditions: []domains.CachePlanCondition{},
							Orders:     []domains.CachePlanOrder{},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "SELECT * FROM users WHERE id = ?;",
							Type:  domains.CachePlanQueryType_SELECT,
						},
						Select: &domains.CachePlanSelectQuery{
							Cache:   true,
							Table:   "users",
							Targets: []string{"created_at", "id", "name", "username"},
							Conditions: []domains.CachePlanCondition{
								{Column: "id", Operator: domains.CachePlanOperator_EQ, Placeholder: domains.CachePlanPlaceholder{Index: 0}},
							},
							Orders: []domains.CachePlanOrder{},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "SELECT * FROM users WHERE id = ? AND name = ?;",
							Type:  domains.CachePlanQueryType_SELECT,
						},
						Select: &domains.CachePlanSelectQuery{
							Cache:   true,
							Table:   "users",
							Targets: []string{"created_at", "id", "name", "username"},
							Conditions: []domains.CachePlanCondition{
								{Column: "id", Operator: domains.CachePlanOperator_EQ, Placeholder: domains.CachePlanPlaceholder{Index: 0}},
								{Column: "name", Operator: domains.CachePlanOperator_EQ, Placeholder: domains.CachePlanPlaceholder{Index: 1}},
							},
							Orders: []domains.CachePlanOrder{},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "SELECT * FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?;",
							Type:  domains.CachePlanQueryType_SELECT,
						},
						Select: &domains.CachePlanSelectQuery{
							Cache:   true,
							Table:   "users",
							Targets: []string{"created_at", "id", "name", "username"},
							Conditions: []domains.CachePlanCondition{
								{Column: "OFFSET()", Operator: domains.CachePlanOperator_EQ, Placeholder: domains.CachePlanPlaceholder{Index: 0}},
								{Column: "LIMIT()", Operator: domains.CachePlanOperator_EQ, Placeholder: domains.CachePlanPlaceholder{Index: 0, Extra: true}},
							},
							Orders: []domains.CachePlanOrder{
								{Column: "created_at", Order: domains.CachePlanOrder_DESC},
							},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "UPDATE users SET name = ? WHERE id = ?;",
							Type:  domains.CachePlanQueryType_UPDATE,
						},
						Update: &domains.CachePlanUpdateQuery{
							Table:      "users",
							Targets:    []domains.CachePlanUpdateTarget{{Column: "name", Placeholder: domains.CachePlanPlaceholder{Index: 0}}},
							Conditions: []domains.CachePlanCondition{{Column: "id", Operator: domains.CachePlanOperator_EQ, Placeholder: domains.CachePlanPlaceholder{Index: 1}}},
							Orders:     []domains.CachePlanOrder{},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "UPDATE users SET name = ? WHERE id IN (?);",
							Type:  domains.CachePlanQueryType_UPDATE,
						},
						Update: &domains.CachePlanUpdateQuery{
							Table:      "users",
							Targets:    []domains.CachePlanUpdateTarget{{Column: "name", Placeholder: domains.CachePlanPlaceholder{Index: 0, Extra: true}}},
							Conditions: []domains.CachePlanCondition{{Column: "id", Operator: domains.CachePlanOperator_IN, Placeholder: domains.CachePlanPlaceholder{Index: 0}}},
							Orders:     []domains.CachePlanOrder{},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "DELETE FROM users WHERE id = ?;",
							Type:  domains.CachePlanQueryType_DELETE,
						},
						Delete: &domains.CachePlanDeleteQuery{
							Table:      "users",
							Conditions: []domains.CachePlanCondition{{Column: "id", Operator: domains.CachePlanOperator_EQ, Placeholder: domains.CachePlanPlaceholder{Index: 0}}},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "DELETE FROM users WHERE id IN (?);",
							Type:  domains.CachePlanQueryType_DELETE,
						},
						Delete: &domains.CachePlanDeleteQuery{
							Table:      "users",
							Conditions: []domains.CachePlanCondition{{Column: "id", Operator: domains.CachePlanOperator_IN, Placeholder: domains.CachePlanPlaceholder{Index: 0}}},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "INSERT INTO users (name, username, created_at) VALUES (?);",
							Type:  domains.CachePlanQueryType_INSERT,
						},
						Insert: &domains.CachePlanInsertQuery{
							Table:   "users",
							Columns: []string{"name", "username", "created_at"},
						},
					},
				},
			},
		},
		{
			name: "private-isu",
			queries: []string{
				"DELETE FROM `users` WHERE `id` > 1000;",
				"SELECT COUNT(*) AS `count` FROM `comments` WHERE `post_id` IN (?);",
				"SELECT COUNT(*) AS `count` FROM `comments` WHERE `post_id` = ?;",
				"UPDATE `users` SET `del_flg` = 1 WHERE `id` % 50 = 0;",
				"SELECT `id`, `user_id`, `body`, `mime`, `created_at` FROM `posts` WHERE `user_id` = ? ORDER BY `created_at` DESC;",
				"INSERT INTO `comments` (`post_id`, `user_id`, `comment`) VALUES (?);",
				"UPDATE `users` SET `del_flg` = 0;",
				"SELECT * FROM `comments` WHERE `post_id` = ? ORDER BY `created_at` DESC;",
				"SELECT * FROM `users` WHERE `id` = ?;",
				"SELECT `id`, `user_id`, `body`, `mime`, `created_at` FROM `posts` WHERE `created_at` <= ? ORDER BY `created_at` DESC;",
				"INSERT INTO `posts` (`user_id`, `mime`, `imgdata`, `body`) VALUES (?);",
				"SELECT 1 FROM `users` WHERE `account_name` = ?;",
				"SELECT `id`, `user_id`, `body`, `mime`, `created_at` FROM `posts` ORDER BY `created_at` DESC;",
				"SELECT COUNT(*) AS `count` FROM `comments` WHERE `post_id` = ?;",
				"INSERT INTO `users` (`account_name`, `passhash`) VALUES (?);",
				"SELECT * FROM `users` WHERE `account_name` = ? AND `del_flg` = 0;",
				"SELECT * FROM `posts` WHERE `id` = ?;",
				"DELETE FROM `posts` WHERE `id` > 10000;",
				"DELETE FROM `comments` WHERE `id` > 100000;",
				"SELECT * FROM `comments` WHERE `post_id` = ? ORDER BY `created_at` DESC LIMIT 3;",
				"SELECT `id` FROM `posts` WHERE `user_id` = ?;",
			},
			schemas: []domains.TableSchema{
				{
					TableName: "users",
					Columns: map[string]domains.TableSchemaColumn{
						"id":           {ColumnName: "id", DataType: domains.TableSchemaDataType_INT, IsNullable: false, IsPrimary: true, IsUnique: false},
						"account_name": {ColumnName: "account_name", DataType: domains.TableSchemaDataType_STRING, IsNullable: false, IsPrimary: false, IsUnique: true},
						"passhash":     {ColumnName: "passhash", DataType: domains.TableSchemaDataType_STRING, IsNullable: false, IsPrimary: false, IsUnique: false},
						"authority":    {ColumnName: "authority", DataType: domains.TableSchemaDataType_INT, IsNullable: false, IsPrimary: false, IsUnique: false},
						"del_flg":      {ColumnName: "del_flg", DataType: domains.TableSchemaDataType_INT, IsNullable: false, IsPrimary: false, IsUnique: false},
						"created_at":   {ColumnName: "created_at", DataType: domains.TableSchemaDataType_DATETIME, IsNullable: false, IsPrimary: false, IsUnique: false},
					},
				},
				{
					TableName: "posts",
					Columns: map[string]domains.TableSchemaColumn{
						"id":         {ColumnName: "id", DataType: domains.TableSchemaDataType_INT, IsNullable: false, IsPrimary: true, IsUnique: false},
						"user_id":    {ColumnName: "user_id", DataType: domains.TableSchemaDataType_INT, IsNullable: false, IsPrimary: false, IsUnique: false},
						"mime":       {ColumnName: "mime", DataType: domains.TableSchemaDataType_STRING, IsNullable: false, IsPrimary: false, IsUnique: false},
						"imgdata":    {ColumnName: "imgdata", DataType: domains.TableSchemaDataType_BYTES, IsNullable: false, IsPrimary: false, IsUnique: false},
						"body":       {ColumnName: "body", DataType: domains.TableSchemaDataType_STRING, IsNullable: false, IsPrimary: false, IsUnique: false},
						"created_at": {ColumnName: "created_at", DataType: domains.TableSchemaDataType_DATETIME, IsNullable: false, IsPrimary: false, IsUnique: false},
					},
				},
				{
					TableName: "comments",
					Columns: map[string]domains.TableSchemaColumn{
						"id":         {ColumnName: "id", DataType: domains.TableSchemaDataType_INT, IsNullable: false, IsPrimary: true, IsUnique: false},
						"post_id":    {ColumnName: "post_id", DataType: domains.TableSchemaDataType_INT, IsNullable: false, IsPrimary: false, IsUnique: false},
						"user_id":    {ColumnName: "user_id", DataType: domains.TableSchemaDataType_INT, IsNullable: false, IsPrimary: false, IsUnique: false},
						"comment":    {ColumnName: "comment", DataType: domains.TableSchemaDataType_STRING, IsNullable: false, IsPrimary: false, IsUnique: false},
						"created_at": {ColumnName: "created_at", DataType: domains.TableSchemaDataType_DATETIME, IsNullable: false, IsPrimary: false, IsUnique: false},
					},
				},
			},
			expected: domains.CachePlan{
				Queries: []*domains.CachePlanQuery{
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "DELETE FROM users WHERE id > 1000;",
							Type:  domains.CachePlanQueryType_DELETE,
						},
						Delete: &domains.CachePlanDeleteQuery{
							Table:      "users",
							Conditions: []domains.CachePlanCondition{},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "SELECT COUNT(*) FROM comments WHERE post_id IN (?);",
							Type:  domains.CachePlanQueryType_SELECT,
						},
						Select: &domains.CachePlanSelectQuery{
							Cache:   true,
							Table:   "comments",
							Targets: []string{"COUNT()"},
							Conditions: []domains.CachePlanCondition{
								{Column: "post_id", Operator: domains.CachePlanOperator_IN, Placeholder: domains.CachePlanPlaceholder{Index: 0}},
							},
							Orders: []domains.CachePlanOrder{},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "SELECT COUNT(*) FROM comments WHERE post_id = ?;",
							Type:  domains.CachePlanQueryType_SELECT,
						},
						Select: &domains.CachePlanSelectQuery{
							Cache:   true,
							Table:   "comments",
							Targets: []string{"COUNT()"},
							Conditions: []domains.CachePlanCondition{
								{Column: "post_id", Operator: domains.CachePlanOperator_EQ, Placeholder: domains.CachePlanPlaceholder{Index: 0}},
							},
							Orders: []domains.CachePlanOrder{},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "UPDATE users SET del_flg = 1 WHERE id % 50 = 0;",
							Type:  domains.CachePlanQueryType_UPDATE,
						},
						Update: &domains.CachePlanUpdateQuery{
							Table: "users",
							Targets: []domains.CachePlanUpdateTarget{
								{Column: "account_name", Placeholder: domains.CachePlanPlaceholder{Index: 0}},
								{Column: "authority", Placeholder: domains.CachePlanPlaceholder{Index: 1}},
								{Column: "created_at", Placeholder: domains.CachePlanPlaceholder{Index: 2}},
								{Column: "del_flg", Placeholder: domains.CachePlanPlaceholder{Index: 3}},
								{Column: "id", Placeholder: domains.CachePlanPlaceholder{Index: 4}},
								{Column: "passhash", Placeholder: domains.CachePlanPlaceholder{Index: 5}},
							},
							Conditions: []domains.CachePlanCondition{},
							Orders:     []domains.CachePlanOrder{},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "SELECT id, user_id, body, mime, created_at FROM posts WHERE user_id = ? ORDER BY created_at DESC;",
							Type:  domains.CachePlanQueryType_SELECT,
						},
						Select: &domains.CachePlanSelectQuery{
							Cache:   true,
							Table:   "posts",
							Targets: []string{"body", "created_at", "id", "mime", "user_id"},
							Conditions: []domains.CachePlanCondition{
								{Column: "user_id", Operator: domains.CachePlanOperator_EQ, Placeholder: domains.CachePlanPlaceholder{Index: 0}},
							},
							Orders: []domains.CachePlanOrder{
								{Column: "created_at", Order: domains.CachePlanOrder_DESC},
							},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "INSERT INTO comments (post_id, user_id, comment) VALUES (?);",
							Type:  domains.CachePlanQueryType_INSERT,
						},
						Insert: &domains.CachePlanInsertQuery{
							Table:   "comments",
							Columns: []string{"post_id", "user_id", "comment"},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "UPDATE users SET del_flg = ?;",
							Type:  domains.CachePlanQueryType_UPDATE,
						},
						Update: &domains.CachePlanUpdateQuery{
							Table:      "users",
							Targets:    []domains.CachePlanUpdateTarget{{Column: "del_flg", Placeholder: domains.CachePlanPlaceholder{Index: 0, Extra: true}}},
							Conditions: []domains.CachePlanCondition{},
							Orders:     []domains.CachePlanOrder{},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "SELECT * FROM comments WHERE post_id = ? ORDER BY created_at DESC;",
							Type:  domains.CachePlanQueryType_SELECT,
						},
						Select: &domains.CachePlanSelectQuery{
							Cache:   true,
							Table:   "comments",
							Targets: []string{"comment", "created_at", "id", "post_id", "user_id"},
							Conditions: []domains.CachePlanCondition{
								{Column: "post_id", Operator: domains.CachePlanOperator_EQ, Placeholder: domains.CachePlanPlaceholder{Index: 0}},
							},
							Orders: []domains.CachePlanOrder{
								{Column: "created_at", Order: domains.CachePlanOrder_DESC},
							},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "SELECT * FROM users WHERE id = ?;",
							Type:  domains.CachePlanQueryType_SELECT,
						},
						Select: &domains.CachePlanSelectQuery{
							Cache:   true,
							Table:   "users",
							Targets: []string{"account_name", "authority", "created_at", "del_flg", "id", "passhash"},
							Conditions: []domains.CachePlanCondition{
								{Column: "id", Operator: domains.CachePlanOperator_EQ, Placeholder: domains.CachePlanPlaceholder{Index: 0}},
							},
							Orders: []domains.CachePlanOrder{},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "SELECT id, user_id, body, mime, created_at FROM posts WHERE created_at <= ? ORDER BY created_at DESC;",
							Type:  domains.CachePlanQueryType_SELECT,
						},
						Select: &domains.CachePlanSelectQuery{
							Cache: false,
							Table: "posts",
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "INSERT INTO posts (user_id, mime, imgdata, body) VALUES (?);",
							Type:  domains.CachePlanQueryType_INSERT,
						},
						Insert: &domains.CachePlanInsertQuery{
							Table:   "posts",
							Columns: []string{"user_id", "mime", "imgdata", "body"},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "SELECT 1 FROM users WHERE account_name = ?;",
							Type:  domains.CachePlanQueryType_SELECT,
						},
						Select: &domains.CachePlanSelectQuery{
							Cache:   true,
							Table:   "users",
							Targets: []string{},
							Conditions: []domains.CachePlanCondition{
								{Column: "account_name", Operator: domains.CachePlanOperator_EQ, Placeholder: domains.CachePlanPlaceholder{Index: 0}},
							},
							Orders: []domains.CachePlanOrder{},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "SELECT id, user_id, body, mime, created_at FROM posts ORDER BY created_at DESC;",
							Type:  domains.CachePlanQueryType_SELECT,
						},
						Select: &domains.CachePlanSelectQuery{
							Cache:      true,
							Table:      "posts",
							Targets:    []string{"body", "created_at", "id", "mime", "user_id"},
							Conditions: []domains.CachePlanCondition{},
							Orders: []domains.CachePlanOrder{
								{Column: "created_at", Order: domains.CachePlanOrder_DESC},
							},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "SELECT COUNT(*) FROM comments WHERE post_id = ?;",
							Type:  domains.CachePlanQueryType_SELECT,
						},
						Select: &domains.CachePlanSelectQuery{
							Cache:   true,
							Table:   "comments",
							Targets: []string{"COUNT()"},
							Conditions: []domains.CachePlanCondition{
								{Column: "post_id", Operator: domains.CachePlanOperator_EQ, Placeholder: domains.CachePlanPlaceholder{Index: 0}},
							},
							Orders: []domains.CachePlanOrder{},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "INSERT INTO users (account_name, passhash) VALUES (?);",
							Type:  domains.CachePlanQueryType_INSERT,
						},
						Insert: &domains.CachePlanInsertQuery{
							Table:   "users",
							Columns: []string{"account_name", "passhash"},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "SELECT * FROM users WHERE account_name = ? AND del_flg = ?;",
							Type:  domains.CachePlanQueryType_SELECT,
						},
						Select: &domains.CachePlanSelectQuery{
							Cache:   true,
							Table:   "users",
							Targets: []string{"account_name", "authority", "created_at", "del_flg", "id", "passhash"},
							Conditions: []domains.CachePlanCondition{
								{Column: "account_name", Operator: domains.CachePlanOperator_EQ, Placeholder: domains.CachePlanPlaceholder{Index: 0}},
								{Column: "del_flg", Operator: domains.CachePlanOperator_EQ, Placeholder: domains.CachePlanPlaceholder{Index: 0, Extra: true}},
							},
							Orders: []domains.CachePlanOrder{},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "SELECT * FROM posts WHERE id = ?;",
							Type:  domains.CachePlanQueryType_SELECT,
						},
						Select: &domains.CachePlanSelectQuery{
							Cache:   true,
							Table:   "posts",
							Targets: []string{"body", "created_at", "id", "imgdata", "mime", "user_id"},
							Conditions: []domains.CachePlanCondition{
								{Column: "id", Operator: domains.CachePlanOperator_EQ, Placeholder: domains.CachePlanPlaceholder{Index: 0}},
							},
							Orders: []domains.CachePlanOrder{},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "DELETE FROM posts WHERE id > 10000;",
							Type:  domains.CachePlanQueryType_DELETE,
						},
						Delete: &domains.CachePlanDeleteQuery{
							Table:      "posts",
							Conditions: []domains.CachePlanCondition{},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "DELETE FROM comments WHERE id > 100000;",
							Type:  domains.CachePlanQueryType_DELETE,
						},
						Delete: &domains.CachePlanDeleteQuery{
							Table:      "comments",
							Conditions: []domains.CachePlanCondition{},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "SELECT * FROM comments WHERE post_id = ? ORDER BY created_at DESC LIMIT ?;",
							Type:  domains.CachePlanQueryType_SELECT,
						},
						Select: &domains.CachePlanSelectQuery{
							Cache:   true,
							Table:   "comments",
							Targets: []string{"comment", "created_at", "id", "post_id", "user_id"},
							Conditions: []domains.CachePlanCondition{
								{Column: "post_id", Operator: domains.CachePlanOperator_EQ, Placeholder: domains.CachePlanPlaceholder{Index: 0}},
								{Column: "LIMIT()", Operator: domains.CachePlanOperator_EQ, Placeholder: domains.CachePlanPlaceholder{Index: 0, Extra: true}},
							},
							Orders: []domains.CachePlanOrder{
								{Column: "created_at", Order: domains.CachePlanOrder_DESC},
							},
						},
					},
					{
						CachePlanQueryBase: &domains.CachePlanQueryBase{
							Query: "SELECT id FROM posts WHERE user_id = ?;",
							Type:  domains.CachePlanQueryType_SELECT,
						},
						Select: &domains.CachePlanSelectQuery{
							Cache:   true,
							Table:   "posts",
							Targets: []string{"id"},
							Conditions: []domains.CachePlanCondition{
								{Column: "user_id", Operator: domains.CachePlanOperator_EQ, Placeholder: domains.CachePlanPlaceholder{Index: 0}},
							},
							Orders: []domains.CachePlanOrder{},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := AnalyzeQueries(test.queries, test.schemas)
			assert.NoError(t, err)
			fmt.Printf("%+v\n", actual)
			assert.Equal(t, test.expected, actual)
		})
	}
}
