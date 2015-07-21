package schemabuilder

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

var P = fmt.Println

var PtrString *string

type Person struct {
	Id        uint64 `pk:"true" autoincrement:"true"`
	Name      string `size:"200" unique:"true"`
	Info      *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Student struct {
	Person
	ClassRoom string
	Grade     int32
}

type PersonIndex struct {
	key1 IComposedKeys `columns:"CreatedAt,UpdatedAt"`
	key2 IUniqueKeys   `columns:"CreatedAt,UpdatedAt"`
}

type TestTable_v1 struct {
	A string
	B string
	C int64 `unique:"true"`
}
type TestTable_v2 struct {
	TestTable_v1
	D string
}
type TestTable_v3 struct {
	A string `size:"1000000"`
	B string
	D int64
}
type TestTable TestTable_v3

func typeOf(v interface{}) reflect.Type {
	return reflect.TypeOf(v)
}

func replace(s, o, n string) string {
	return strings.Replace(s, o, n, -1)
}

var builderMySQL = For(NewMySQLDialect("utf8", "InnoDB"))
var builderSQLite = For(NewSQLite3Dialect())
var builderPostgres = For(NewPostgresDialect())

func TestVersioning(t *testing.T) {
	table := builderMySQL.DefineTable(TestTable{}, nil).
		AddHistory(TestTable_v1{}, nil).
		AddHistory(TestTable_v2{}, nil).
		AddHistory(TestTable_v3{}, nil)
	P(table.Histories[0])
	P(table.MigrateSQL("v1", "current"))
}

func TestEmbedStructMySQL(t *testing.T) {

	expect := replace(`CREATE TABLE IF NOT EXISTS +students+(
+id+ BIGINT AUTO_INCREMENT NOT NULL PRIMARY KEY,
+name+ VARCHAR(200) NOT NULL UNIQUE,
+info+ VARCHAR(255) ,
+created_at+ DATETIME NOT NULL,
+updated_at+ DATETIME NOT NULL,
+class_room+ VARCHAR(255) NOT NULL,
+grade+ INT NOT NULL
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8;
CREATE INDEX +key1+ ON +students+ (+created_at+,+updated_at+);
CREATE UNIQUE INDEX +key2+ ON +students+ (+created_at+,+updated_at+);
`, "+", "`")
	table := builderMySQL.DefineTable(Student{}, PersonIndex{})
	if result := table.String(); result != expect {
		t.Errorf("invalid create table %s", result)
	}
}

func TestSample02MySQL(t *testing.T) {
	table := builderMySQL.DefineTable(Person{}, PersonIndex{})
	expect := replace(`CREATE TABLE IF NOT EXISTS +people+(
+id+ BIGINT AUTO_INCREMENT NOT NULL PRIMARY KEY,
+name+ VARCHAR(200) NOT NULL UNIQUE,
+info+ VARCHAR(255) ,
+created_at+ DATETIME NOT NULL,
+updated_at+ DATETIME NOT NULL
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8;
CREATE INDEX +key1+ ON +people+ (+created_at+,+updated_at+);
CREATE UNIQUE INDEX +key2+ ON +people+ (+created_at+,+updated_at+);
`, "+", "`")

	if result := table.String(); result != expect {
		t.Errorf("invalid create table %s", result)
	}
}

func TestSample02SQLite3(t *testing.T) {
	table := builderSQLite.DefineTable(Person{}, PersonIndex{})
	expect := `CREATE TABLE IF NOT EXISTS "people"(
"id" integer AUTOINCREMENT NOT NULL PRIMARY KEY,
"name" text NOT NULL UNIQUE,
"info" text ,
"created_at" datetime NOT NULL,
"updated_at" datetime NOT NULL
);
CREATE INDEX "key1" ON "people" ("created_at","updated_at");
CREATE UNIQUE INDEX "key2" ON "people" ("created_at","updated_at");
`

	if result := table.String(); result != expect {
		t.Errorf("invalid create table %s", result)
	}
}

func TestSample02Postgres(t *testing.T) {
	table := builderPostgres.DefineTable(Person{}, PersonIndex{})
	expect := `CREATE TABLE IF NOT EXISTS "people"(
"id" BIGSERIAL  NOT NULL PRIMARY KEY,
"name" VARCHAR(200) NOT NULL UNIQUE,
"info" VARCHAR(255) ,
"created_at" TIMESTAMP WITH TIME ZONE NOT NULL,
"updated_at" TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE INDEX "key1" ON "people" ("created_at","updated_at");
CREATE UNIQUE INDEX "key2" ON "people" ("created_at","updated_at");
`

	if result := table.String(); result != expect {
		t.Errorf("invalid create table \n\t%s\n\t%s", result, expect)
	}
}
