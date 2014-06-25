package schemabuilder

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type MySQLDialect struct {
	Charset string
	Engine  string
}

func (d *MySQLDialect) Column() ColumnMapper {
	return &MySQLColumnMapper{d}
}

func (d *MySQLDialect) Quote(s string) string {
	return fmt.Sprintf("`%s`", strings.Replace(s, "`", "``", -1))
}

func (d *MySQLDialect) CreateTableSuffix() string {
	suffix := make([]string, 0)
	if d.Engine != "" {
		suffix = append(suffix, fmt.Sprintf("ENGINE=%s", d.Engine))
	}
	if d.Charset != "" {
		suffix = append(suffix, fmt.Sprintf("DEFAULT CHARACTER SET=%s", d.Charset))
	}
	return fmt.Sprintf(" %s", strings.Join(suffix, " "))
}

func NewMySQLDialect(charset string, engine string) *MySQLDialect {
	return &MySQLDialect{Charset: charset, Engine: engine}
}

type MySQLColumnMapper struct {
	Dialect *MySQLDialect
}

func (m *MySQLColumnMapper) Quote(s string) string {
	return m.Dialect.Quote(s)
}

func (m *MySQLColumnMapper) DataType(v interface{}, size uint64) string {
	switch v.(type) {
	case bool:
		return "BOOLEAN"
	case *bool, sql.NullBool:
		return "BOOLEAN"
	case int8, int16, uint8, uint16:
		return "SMALLINT"
	case *int8, *int16, *uint8, *uint16:
		return "SMALLINT"
	case int, int32, uint, uint32:
		return "INT"
	case *int, *int32, *uint, *uint32:
		return "INT"
	case int64, uint64:
		return "BIGINT"
	case *int64, *uint64, sql.NullInt64:
		return "BIGINT"
	case string:
		return m.varchar(size)
	case *string, sql.NullString:
		return m.varchar(size)
	case []byte:
		switch {
		case size == 0:
			return "VARBINARY(255)" // default.
		case size < 65533: // approximate 64KB.
			return fmt.Sprintf("VARBINARY(%d)", size)
		case size < 1<<24: // 16MB.
			return "MEDIUMBLOB"
		}
		return "LONGBLOB"
	case time.Time:
		return "DATETIME"
	case *time.Time:
		return "DATETIME"
	case float32, *float32, float64, *float64, sql.NullFloat64:
		return "DOUBLE"
	}
	panic(fmt.Errorf("MySQLDialect: unsupported SQL type: %T", v))
}

func (d *MySQLColumnMapper) varchar(size uint64) string {
	switch {
	case size == 0:
		return "VARCHAR(255)" // default.
	case size < (1<<16)-1-2: // approximate 64KB.
		return fmt.Sprintf("VARCHAR(%d)", size)
	case size < 1<<24: // 16MB.
		return "MEDIUMTEXT"
	}
	return "LONGTEXT"
}

func (m *MySQLColumnMapper) AutoIncrement() string {
	return "AUTO_INCREAMENT"
}
