package domains

type TableSchema struct {
	TableName string
	// ColumnName -> ColumnSchema
	Columns map[string]TableSchemaColumn
}

func newTableSchema() TableSchema {
	return TableSchema{
		Columns: make(map[string]TableSchemaColumn),
	}
}

type TableSchemaColumn struct {
	ColumnName string
	DataType   TableSchemaDataType
	IsNullable bool
	IsPrimary  bool
	IsUnique   bool
}

type TableSchemaDataType string

const (
	TableSchemaDataType_STRING   TableSchemaDataType = "string"
	TableSchemaDataType_BYTES    TableSchemaDataType = "bytes"
	TableSchemaDataType_INT      TableSchemaDataType = "int"
	TableSchemaDataType_INT64    TableSchemaDataType = "int64"
	TableSchemaDataType_DATETIME TableSchemaDataType = "time"
	TableSchemaDataType_UNKNOWN  TableSchemaDataType = "unknown"
)
