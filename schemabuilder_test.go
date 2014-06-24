package schemabuilder

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

var P = fmt.Println

var Int64 int64
var Int32 int32
var Int16 int16
var Int8 int8
var Bool bool
var String string
var PtrString *string

type Person struct {
	Id        uint64 `pk:"true",autoincrement:"true"`
	Name      string `size:"200",unique:"true"`
	Info      *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PersonIndex struct {
	key1 IComposedKeys `columns:"CreatedAt,UpdatedAt"`
	key2 IUniqueKeys   `columns:"CreatedAt,UpdatedAt"`
}

func typeOf(v interface{}) reflect.Type {
	return reflect.TypeOf(v)
}

func GetColumnMapper(driver string) ColumnMapper {
	if driver == "sqlite3" {
		return &SQLite3ColumnMapper{}
	}
	if driver == "mysql" {
		return &MySQLColumnMapper{}
	}
	panic("cannot find type mapper")
}

func TestSample02MySQL(t *testing.T) {
	table := CreateTableByStruct(Person{}, PersonIndex{}, NewMySQLDialect("utf8", "InnoDB"))
	expect := `CREATE TABLE IF NOT EXISTS people(
id BIGINT NOT NULL PRIMARY KEY,
name VARCHAR(200) NOT NULL,
info VARCHAR(255) ,
created_at DATETIME NOT NULL,
updated_at DATETIME NOT NULL
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8;
CREATE INDEX key1 ON people (created_at,updated_at);
CREATE UNIQUE INDEX key2 ON people (created_at,updated_at);
`

	if result := table.String(); result != expect {
		t.Errorf("invalid create table %s", result)
	}
}

func TestSample02SQLite3(t *testing.T) {
	table := CreateTableByStruct(Person{}, PersonIndex{}, NewSQLite3Dialect())
	expect := `CREATE TABLE IF NOT EXISTS people(
id integer NOT NULL PRIMARY KEY,
name text NOT NULL,
info text ,
created_at datetime NOT NULL,
updated_at datetime NOT NULL
);
CREATE INDEX key1 ON people (created_at,updated_at);
CREATE UNIQUE INDEX key2 ON people (created_at,updated_at);
`

	if result := table.String(); result != expect {
		t.Errorf("invalid create table %s", result)
	}
}

func TestSample01MySQL(t *testing.T) {
	expect := "name MEDIUMTEXT DEFAULT 'hello' AUTO_INCREAMENT PRIMARY KEY"
	column := NewColumn("name").
		As(PtrString).
		Default("'hello'").
		AutoIncrement().
		PrimaryKey().
		Size(1000000)

	dialect := GetColumnMapper("mysql")
	if result := column.ToSQL(dialect); result != expect {
		t.Errorf("SQL should be %s but %s", expect, result)
	}
}

func TestSample01SQLite3(t *testing.T) {
	expect := "name text DEFAULT 'hello' AUTOINCREAMENT PRIMARY KEY"
	column := NewColumn("name").
		As(PtrString).
		Default("'hello'").
		AutoIncrement().
		PrimaryKey().
		Size(1000000)

	dialect := GetColumnMapper("sqlite3")
	if result := column.ToSQL(dialect); result != expect {
		t.Errorf("SQL should be \n'%s' but \n'%s'", expect, result)
	}
}
