package domains

import (
	"regexp"
	"strings"
)

var tableDefRegex = regexp.MustCompile(`(?s)CREATE TABLE ` + "`?" + `(?P<TableName>\w+)` + "`?" + ` \((?P<Columns>.*?)\);`)
var columnDefRegex = regexp.MustCompile(`^` + "`?" + `(?P<ColumnName>\w+)` + "`?" + `\s+(?P<DataType>\w+(?:\([^)]*\))?)(?:(?P<AutoIncrement>\s+AUTO_INCREMENT)|(?P<Unique>\s+UNIQUE)|(?P<PrimaryKey>\s+PRIMARY KEY)|(?P<NonNullable>\s+NOT NULL))*$`)
var primaryKeyRegex = regexp.MustCompile(`^PRIMARY KEY \(` + "`?" + `(?P<ColumnName>\w+)` + "`?" + `\)$`)
var uniqueRegex = regexp.MustCompile(`^UNIQUE(?: KEY)?\s+(?:` + "`?" + `\w+` + "`?" + `\s+)?\(` + "`?" + `(?P<ColumnName>\w+)` + "`?" + `\)$`)

func LoadTableSchema(sql string) ([]TableSchema, error) {
	var tableSchemas []TableSchema

	tableMatches := tableDefRegex.FindAllStringSubmatch(sql, -1)
	tableNames := tableDefRegex.SubexpNames()

	for _, tableMatch := range tableMatches {
		var schema TableSchema
		for i, name := range tableNames {
			if name == "TableName" {
				schema.TableName = tableMatch[i]
			} else if name == "Columns" {
				columnsDef := tableMatch[i]
				columnDefs := strings.Split(columnsDef, ",")
				for _, columnDef := range columnDefs {
					columnDef = strings.TrimSpace(columnDef)
					primaryKeyMatches := primaryKeyRegex.FindStringSubmatch(columnDef)
					if len(primaryKeyMatches) > 0 {
						primaryKeyNames := primaryKeyRegex.SubexpNames()
						for j, name := range primaryKeyNames {
							if name == "ColumnName" {
								for i, column := range schema.Columns {
									if column.ColumnName == primaryKeyMatches[j] {
										schema.Columns[i].IsPrimary = true
									}
								}
							}
						}
					}

					uniqueMatches := uniqueRegex.FindStringSubmatch(columnDef)
					if len(uniqueMatches) > 0 {
						uniqueNames := uniqueRegex.SubexpNames()
						for j, name := range uniqueNames {
							if name == "ColumnName" {
								for i, column := range schema.Columns {
									if column.ColumnName == uniqueMatches[j] {
										schema.Columns[i].IsUnique = true
									}
								}
							}
						}
					}

					columnMatches := columnDefRegex.FindAllStringSubmatch(columnDef, -1)
					columnNames := columnDefRegex.SubexpNames()

					for _, columnMatch := range columnMatches {
						var column TableSchemaColumn
						column.IsNullable = true
						for j, cname := range columnNames {
							switch cname {
							case "ColumnName":
								column.ColumnName = columnMatch[j]
							case "DataType":
								column.DataType = parseDataType(columnMatch[j])
							case "NonNullable":
								column.IsNullable = column.IsNullable && columnMatch[j] == ""
							case "PrimaryKey":
								column.IsPrimary = columnMatch[j] != ""
							case "Unique":
								column.IsUnique = columnMatch[j] != ""
							}
						}
						schema.Columns = append(schema.Columns, column)
					}
				}
			}
		}
		tableSchemas = append(tableSchemas, schema)
	}

	return tableSchemas, nil
}

func parseDataType(sqlType string) TableSchemaDataType {
	sqlType = strings.ToLower(sqlType)
	sqlType = strings.Split(sqlType, "(")[0]
	switch sqlType {
	case "varchar", "text":
		return TableSchemaDataType_STRING
	case "longblob":
		return TableSchemaDataType_BYTES
	case "int":
		return TableSchemaDataType_INT
	case "bigint":
		return TableSchemaDataType_INT64
	case "time", "date", "datetime":
		return TableSchemaDataType_DATETIME
	default:
		return TableSchemaDataType_UNKNOWN
	}
}
