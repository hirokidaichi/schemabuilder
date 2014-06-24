package schemabuilder

import (
	"database/sql"
	"fmt"
	"time"
)

type SQLite3Dialect struct{}

func NewSQLite3Dialect() *SQLite3Dialect {
	return &SQLite3Dialect{}
}

func (d *SQLite3Dialect) Column() ColumnMapper {
	return &SQLite3ColumnMapper{}
}

func (d *SQLite3Dialect) CreateTableSuffix() string {
	return ""
}

type SQLite3ColumnMapper struct{}

func (m *SQLite3ColumnMapper) DataType(v interface{}, size uint64) string {
	switch v.(type) {
	case bool:
		return "boolean"
	case *bool, sql.NullBool:
		return "boolean"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "integer"
	case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64, sql.NullInt64:
		return "integer"
	case string:
		return "text"
	case *string, sql.NullString:
		return "text"
	case []byte:
		return "blob"
	case time.Time:
		return "datetime"
	case *time.Time:
		return "datetime"
	case float32, *float32, float64, *float64, sql.NullFloat64:
		return "real"
	}
	panic(fmt.Errorf("SQLite3Dialect: unsupported SQL type: %T", v))
}

func (m *SQLite3ColumnMapper) AutoIncrement() string {
	return "AUTOINCREAMENT"
}
