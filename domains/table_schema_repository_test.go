package domains

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadTableSchema(t *testing.T) {
	tests := []struct {
		sql      string
		expected []TableSchema
	}{
		{
			sql: "CREATE TABLE users (\n" +
				"  id BIGINT AUTO_INCREMENT PRIMARY KEY,\n" +
				"  `name` VARCHAR(255) NOT NULL,\n" +
				"  created_at DATETIME(6) UNIQUE,\n" +
				"  description TEXT NOT NULL,\n" +
				"  icon LONGBLOB NOT NULL,\n" +
				"  UNIQUE KEY uniq_name (name),\n" +
				");\n" +
				"CREATE TABLE posts (\n" +
				"  id BIGINT NOT NULL AUTO_INCREMENT,\n" +
				"  title VARCHAR(255) NOT NULL,\n" +
				"  content TEXT NOT NULL,\n" +
				"  PRIMARY KEY (id),\n" +
				"  UNIQUE (`title`),\n" +
				");",
			expected: []TableSchema{
				{
					TableName: "users",
					Columns: map[string]TableSchemaColumn{
						"id":          {ColumnName: "id", DataType: TableSchemaDataType_INT64, IsNullable: true, IsPrimary: true, IsUnique: false},
						"name":        {ColumnName: "name", DataType: TableSchemaDataType_STRING, IsNullable: false, IsPrimary: false, IsUnique: true},
						"created_at":  {ColumnName: "created_at", DataType: TableSchemaDataType_DATETIME, IsNullable: true, IsPrimary: false, IsUnique: true},
						"description": {ColumnName: "description", DataType: TableSchemaDataType_STRING, IsNullable: false, IsPrimary: false, IsUnique: false},
						"icon":        {ColumnName: "icon", DataType: TableSchemaDataType_BYTES, IsNullable: false, IsPrimary: false, IsUnique: false},
					},
				},
				{
					TableName: "posts",
					Columns: map[string]TableSchemaColumn{
						"id":      {ColumnName: "id", DataType: TableSchemaDataType_INT64, IsNullable: false, IsPrimary: true, IsUnique: false},
						"title":   {ColumnName: "title", DataType: TableSchemaDataType_STRING, IsNullable: false, IsPrimary: false, IsUnique: true},
						"content": {ColumnName: "content", DataType: TableSchemaDataType_STRING, IsNullable: false, IsPrimary: false, IsUnique: false},
					},
				},
			},
		},
		{
			sql: "DROP TABLE IF EXISTS users;\n" +
				"CREATE TABLE users (\n" +
				"  `id` int NOT NULL AUTO_INCREMENT PRIMARY KEY,\n" +
				"  `account_name` varchar(64) NOT NULL UNIQUE,\n" +
				"  `passhash` varchar(128) NOT NULL, -- SHA2 512 non-binary (hex)\n" +
				"  `authority` tinyint(1) NOT NULL DEFAULT 0,\n" +
				"  `del_flg` tinyint(1) NOT NULL DEFAULT 0,\n" +
				"  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP\n" +
				") DEFAULT CHARSET=utf8mb4;\n" +
				"\n" +
				"DROP TABLE IF EXISTS posts;\n" +
				"CREATE TABLE posts (\n" +
				"  `id` int NOT NULL AUTO_INCREMENT PRIMARY KEY,\n" +
				"  `user_id` int NOT NULL,\n" +
				"  `mime` varchar(64) NOT NULL,\n" +
				"  `imgdata` mediumblob NOT NULL,\n" +
				"  `body` text NOT NULL,\n" +
				"  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP\n" +
				") DEFAULT CHARSET=utf8mb4;\n" +
				"\n" +
				"DROP TABLE IF EXISTS comments;\n" +
				"CREATE TABLE comments (\n" +
				"  `id` int NOT NULL AUTO_INCREMENT PRIMARY KEY,\n" +
				"  `post_id` int NOT NULL,\n" +
				"  `user_id` int NOT NULL,\n" +
				"  `comment` text NOT NULL,\n" +
				"  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP\n" +
				") DEFAULT CHARSET=utf8mb4;\n",
			expected: []TableSchema{
				{
					TableName: "users",
					Columns: map[string]TableSchemaColumn{
						"id":           {ColumnName: "id", DataType: TableSchemaDataType_INT, IsNullable: false, IsPrimary: true, IsUnique: false},
						"account_name": {ColumnName: "account_name", DataType: TableSchemaDataType_STRING, IsNullable: false, IsPrimary: false, IsUnique: true},
						"passhash":     {ColumnName: "passhash", DataType: TableSchemaDataType_STRING, IsNullable: false, IsPrimary: false, IsUnique: false},
						"authority":    {ColumnName: "authority", DataType: TableSchemaDataType_INT, IsNullable: false, IsPrimary: false, IsUnique: false},
						"del_flg":      {ColumnName: "del_flg", DataType: TableSchemaDataType_INT, IsNullable: false, IsPrimary: false, IsUnique: false},
						"created_at":   {ColumnName: "created_at", DataType: TableSchemaDataType_DATETIME, IsNullable: false, IsPrimary: false, IsUnique: false},
					},
				},
				{
					TableName: "posts",
					Columns: map[string]TableSchemaColumn{
						"id":         {ColumnName: "id", DataType: TableSchemaDataType_INT, IsNullable: false, IsPrimary: true, IsUnique: false},
						"user_id":    {ColumnName: "user_id", DataType: TableSchemaDataType_INT, IsNullable: false, IsPrimary: false, IsUnique: false},
						"mime":       {ColumnName: "mime", DataType: TableSchemaDataType_STRING, IsNullable: false, IsPrimary: false, IsUnique: false},
						"imgdata":    {ColumnName: "imgdata", DataType: TableSchemaDataType_BYTES, IsNullable: false, IsPrimary: false, IsUnique: false},
						"body":       {ColumnName: "body", DataType: TableSchemaDataType_STRING, IsNullable: false, IsPrimary: false, IsUnique: false},
						"created_at": {ColumnName: "created_at", DataType: TableSchemaDataType_DATETIME, IsNullable: false, IsPrimary: false, IsUnique: false},
					},
				},
				{
					TableName: "comments",
					Columns: map[string]TableSchemaColumn{
						"id":         {ColumnName: "id", DataType: TableSchemaDataType_INT, IsNullable: false, IsPrimary: true, IsUnique: false},
						"post_id":    {ColumnName: "post_id", DataType: TableSchemaDataType_INT, IsNullable: false, IsPrimary: false, IsUnique: false},
						"user_id":    {ColumnName: "user_id", DataType: TableSchemaDataType_INT, IsNullable: false, IsPrimary: false, IsUnique: false},
						"comment":    {ColumnName: "comment", DataType: TableSchemaDataType_STRING, IsNullable: false, IsPrimary: false, IsUnique: false},
						"created_at": {ColumnName: "created_at", DataType: TableSchemaDataType_DATETIME, IsNullable: false, IsPrimary: false, IsUnique: false},
					},
				},
			},
		},
	}

	for i, test := range tests {
		t.Run("test"+fmt.Sprint(i), func(t *testing.T) {
			schemas, err := LoadTableSchema(test.sql)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, schemas)
		})
	}
}
