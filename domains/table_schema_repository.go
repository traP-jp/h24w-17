package domains

import (
	"regexp"
	"slices"
	"strings"
)

var commentRegex = regexp.MustCompile("(?m)--.*$")
var tableDefRegex = regexp.MustCompile("(?s)CREATE TABLE `?(?P<TableName>\\w+)`? \\((?P<Columns>.*?)\\)[^\\n]*;")
var columnDefRegex = regexp.MustCompile("^`?(?P<ColumnName>\\w+)`?\\s+(?P<DataType>\\w+(?:\\([^)]*\\))?)(?:(?P<AutoIncrement>\\s+AUTO_INCREMENT)|(?P<Unique>\\s+UNIQUE)|(?P<PrimaryKey>\\s+PRIMARY KEY)|(?P<NonNullable>\\s+NOT NULL)|(?P<Default>\\s+DEFAULT [^\\s]+))*$")
var primaryKeyRegex = regexp.MustCompile("^PRIMARY KEY \\(`?(?P<ColumnName>\\w+)`?\\)$")
var uniqueRegex = regexp.MustCompile("^UNIQUE(?: KEY)?\\s+(?:`?\\w+`?\\s+)?\\(`?(?P<ColumnName>\\w+)`?\\)$")

func LoadTableSchema(sql string) ([]TableSchema, error) {
	var tableSchemas []TableSchema

	sql = commentRegex.ReplaceAllString(sql, "")
	tableMatches := tableDefRegex.FindAllStringSubmatch(sql, -1)
	tableNames := tableDefRegex.SubexpNames()

	for _, tableMatch := range tableMatches {
		var schema = newTableSchema()

		var tableName string
		for i, name := range tableNames {
			if name == "TableName" {
				tableName = tableMatch[i]
			}
		}
		schema.TableName = tableName

		for i := range tableNames {
			columnsDef := tableMatch[i]
			columnDefs := strings.Split(columnsDef, ",")
			for _, columnDef := range columnDefs {
				columnDef = strings.TrimSpace(columnDef)

				primaryKeyMatches := primaryKeyRegex.FindStringSubmatch(columnDef)
				if len(primaryKeyMatches) > 0 {
					primaryKeyNames := primaryKeyRegex.SubexpNames()
					j := slices.Index(primaryKeyNames, "ColumnName")
					if j >= 0 {
						primaryKeyColumn, ok := schema.Columns[primaryKeyMatches[j]]
						if ok {
							primaryKeyColumn.IsPrimary = true
							schema.Columns[primaryKeyMatches[j]] = primaryKeyColumn
						}
					}
				}

				uniqueMatches := uniqueRegex.FindStringSubmatch(columnDef)
				if len(uniqueMatches) > 0 {
					uniqueNames := uniqueRegex.SubexpNames()
					j := slices.Index(uniqueNames, "ColumnName")
					if j >= 0 {
						uniqueColumn, ok := schema.Columns[uniqueMatches[j]]
						if ok {
							uniqueColumn.IsUnique = true
							schema.Columns[uniqueMatches[j]] = uniqueColumn
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
					schema.Columns[column.ColumnName] = column
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
	case "longblob", "mediumblob", "blob":
		return TableSchemaDataType_BYTES
	case "int", "tinyint":
		return TableSchemaDataType_INT
	case "bigint":
		return TableSchemaDataType_INT64
	case "time", "date", "datetime", "timestamp":
		return TableSchemaDataType_DATETIME
	default:
		return TableSchemaDataType_UNKNOWN
	}
}
