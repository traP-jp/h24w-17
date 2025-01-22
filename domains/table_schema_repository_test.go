package domains

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadTableSchema(t *testing.T) {
	tests := []struct {
		sql      string
		expected []TableSchema
	}{
		{
			sql: `
			CREATE TABLE users (
				id BIGINT AUTO_INCREMENT PRIMARY KEY,
				` + "`" + `name` + "`" + ` VARCHAR(255) NOT NULL,
				created_at DATETIME(6) UNIQUE,
  			description TEXT NOT NULL,
				icon LONGBLOB NOT NULL,
				UNIQUE KEY uniq_name (name),
			);
			CREATE TABLE posts (
				id BIGINT NOT NULL AUTO_INCREMENT,
				title VARCHAR(255) NOT NULL,
				content TEXT NOT NULL,
				PRIMARY KEY (id),
				UNIQUE (` + "`" + `title` + "`" + `),
			);
			`,
			expected: []TableSchema{
				{
					TableName: "users",
					Columns: []TableSchemaColumn{
						{ColumnName: "id", DataType: TableSchemaDataType_INT64, IsNullable: true, IsPrimary: true, IsUnique: false},
						{ColumnName: "name", DataType: TableSchemaDataType_STRING, IsNullable: false, IsPrimary: false, IsUnique: true},
						{ColumnName: "created_at", DataType: TableSchemaDataType_DATETIME, IsNullable: true, IsPrimary: false, IsUnique: true},
						{ColumnName: "description", DataType: TableSchemaDataType_STRING, IsNullable: false, IsPrimary: false, IsUnique: false},
						{ColumnName: "icon", DataType: TableSchemaDataType_BYTES, IsNullable: false, IsPrimary: false, IsUnique: false},
					},
				},
				{
					TableName: "posts",
					Columns: []TableSchemaColumn{
						{ColumnName: "id", DataType: TableSchemaDataType_INT64, IsNullable: false, IsPrimary: true, IsUnique: false},
						{ColumnName: "title", DataType: TableSchemaDataType_STRING, IsNullable: false, IsPrimary: false, IsUnique: true},
						{ColumnName: "content", DataType: TableSchemaDataType_STRING, IsNullable: false, IsPrimary: false, IsUnique: false},
					},
				},
			},
		},
	}

	for _, test := range tests {
		schemas, err := LoadTableSchema(test.sql)
		assert.NoError(t, err)
		assert.Equal(t, test.expected, schemas)
	}
}
