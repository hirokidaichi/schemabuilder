package schemabuilder

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type PostgresDialect struct{}

func (d *PostgresDialect) Column() ColumnMapper {
	return &PostgresColumnMapper{d}
}

func (d *PostgresDialect) Quote(s string) string {
	return fmt.Sprintf(`"%s"`, strings.Replace(s, `"`, `""`, -1))
}

func (d *PostgresDialect) CreateTableSuffix() string {
	return ""
}

func NewPostgresDialect() *PostgresDialect {
	return &PostgresDialect{}
}

type PostgresColumnMapper struct {
	Dialect *PostgresDialect
}

func (m *PostgresColumnMapper) Quote(s string) string {
	return m.Dialect.Quote(s)
}

func (m *PostgresColumnMapper) DataType(v interface{}, autoIncr bool, size uint64) string {
	switch v.(type) {
	case bool:
		return "BOOLEAN"
	case *bool, sql.NullBool:
		return "BOOLEAN"
	case int8, int16, uint8, uint16:
		return m.smallint(autoIncr)
	case *int8, *int16, *uint8, *uint16:
		return m.smallint(autoIncr)
	case int, int32, uint, uint32:
		return m.integer(autoIncr)
	case *int, *int32, *uint, *uint32:
		return m.integer(autoIncr)
	case int64, uint64:
		return m.bigint(autoIncr)
	case *int64, *uint64, sql.NullInt64:
		return m.bigint(autoIncr)
	case string:
		return m.varchar(size)
	case *string, sql.NullString:
		return m.varchar(size)
	case []byte:
		switch {
		case size == 0:
			return "BIT VARYING(255)" // default.
		case size < 65533: // approximate 64KB.
			return fmt.Sprintf("BIT VARYING(%d)", size)
		case size < 1<<24: // 16MB.
			return "BYTEA"
		}
		return "BYTEA"
	case time.Time:
		return "TIMESTAMP WITH TIME ZONE"
	case *time.Time:
		return "TIMESTAMP WITH TIME ZONE"
	case float32, *float32:
		return "REAL"
	case float64, *float64, sql.NullFloat64:
		return "DOUBLE PRECISION"
	}
	panic(fmt.Errorf("PostgresDialect: unsupported SQL type: %T", v))
}

func (d *PostgresColumnMapper) smallint(autoIncr bool) string {
	if autoIncr {
		return "SMALLSERIAL"
	}
	return "SMALLINT"
}

func (d *PostgresColumnMapper) integer(autoIncr bool) string {
	if autoIncr {
		return "SERIAL"
	}
	return "INTEGER"
}

func (d *PostgresColumnMapper) bigint(autoIncr bool) string {
	if autoIncr {
		return "BIGSERIAL"
	}
	return "BIGINT"
}

func (d *PostgresColumnMapper) varchar(size uint64) string {
	switch {
	case size == 0:
		return "VARCHAR(255)" // default.
	case size < (1<<16)-1-2: // approximate 64KB.
		return fmt.Sprintf("VARCHAR(%d)", size)
	case size < 1<<24: // 16MB.
		return "TEXT"
	}
	return "TEXT"
}

func (m *PostgresColumnMapper) AutoIncrement() string {
	return ""
}
