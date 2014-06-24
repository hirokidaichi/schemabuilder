package schemabuilder

import (
	"bitbucket.org/pkg/inflect"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type IComposedKeys interface{}
type IUniqueKeys interface{}

type Column struct {
	Name            string
	Type            interface{}
	IsNotNull       bool
	DefaultVal      string
	IsAutoIncrement bool
	IsPrimaryKey    bool
	IsUnique        bool
	TypeSize        uint64
}

func NewColumn(name string) *Column {
	return &Column{Name: name, IsNotNull: true}
}

func CreateColumnByField(value interface{}, field reflect.StructField) *Column {
	column := NewColumn(inflect.Underscore(field.Name)).As(value)
	if field.Tag.Get("pk") == "true" {
		column.PrimaryKey()
	}
	if field.Tag.Get("unique") == "true" {
		column.Unique()
	}
	if field.Tag.Get("autoincrement") == "true" {
		column.AutoIncrement()
	}
	if sizeVal := field.Tag.Get("size"); sizeVal != "" {
		n, err := strconv.ParseUint(sizeVal, 10, 64)
		if err != nil {
			panic(err)
		}
		column.Size(n)
	}
	return column
}

func (c *Column) As(t interface{}) *Column {
	c.Type = t
	switch t.(type) {
	case sql.NullBool, sql.NullString, sql.NullFloat64, sql.NullInt64:
		c.IsNotNull = false
	case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64, *bool:
		c.IsNotNull = false
	case *string:
		c.IsNotNull = false
	}
	return c
}

func (c *Column) AutoIncrement() *Column {
	c.IsAutoIncrement = true
	return c
}
func (c *Column) PrimaryKey() *Column {
	c.IsPrimaryKey = true
	return c
}

func (c *Column) Unique() *Column {
	c.IsUnique = true
	return c
}

func (c *Column) Default(v string) *Column {
	c.DefaultVal = v
	return c
}

func (c *Column) Size(v uint64) *Column {
	c.TypeSize = v
	return c
}

func (c *Column) ToSQL(d ColumnMapper) string {
	return fmt.Sprintf("%s %s %s", c.Name, c.DataType(d), c.Constraints(d))
}

func (c *Column) DataType(d ColumnMapper) string {
	return d.DataType(c.Type, c.TypeSize)
}

func (c *Column) Constraints(d ColumnMapper) string {
	options := make([]string, 0)
	if c.DefaultVal != "" {
		options = append(options, "DEFAULT")
		options = append(options, c.DefaultVal)
	}
	if c.IsAutoIncrement {
		options = append(options, d.AutoIncrement())
	}
	if c.IsNotNull {
		options = append(options, "NOT NULL")
	}
	if c.IsPrimaryKey {
		options = append(options, "PRIMARY KEY")
	}
	if c.IsUnique {
		options = append(options, "UNIQUE")
	}
	return strings.Join(options, " ")
}

type Index struct {
	Name        string
	IsUnique    bool
	ColumnNames []string
}

func (i *Index) ToSQL(tableName string) string {
	var format string
	if i.IsUnique {
		format = "CREATE UNIQUE INDEX %s ON %s (%s)"
	} else {
		format = "CREATE INDEX %s ON %s (%s)"
	}
	return fmt.Sprintf(format, i.Name, tableName, strings.Join(i.ColumnNames, ","))
}

func CreateIndexByField(v reflect.StructField) *Index {
	keyName := inflect.Underscore(v.Name)
	typeName := v.Type.Name()
	index := &Index{Name: keyName}
	if typeName == "IUniqueKeys" {
		index.IsUnique = true
	}
	tags := strings.Split(v.Tag.Get("columns"), ",")

	for i, _ := range tags {
		tags[i] = inflect.Underscore(tags[i])
	}
	index.ColumnNames = tags
	return index
}

type Table struct {
	Name    string
	Dialect Dialect
	Columns []*Column
	Indices []*Index
}

func CreateTableByStruct(tableStruct interface{}, indexStruct interface{}, d Dialect) *Table {
	tableName := inflect.Tableize(reflect.TypeOf(tableStruct).Name())
	table := NewTable(tableName, d)
	table.scanColumns(tableStruct)
	table.scanIndices(indexStruct)
	return table
}

func (table *Table) scanColumns(v interface{}) {
	tv := reflect.TypeOf(v)
	vv := reflect.ValueOf(v)
	num := tv.NumField()
	for i := 0; i < num; i++ {
		table.AddColumn(CreateColumnByField(vv.Field(i).Interface(), tv.Field(i)))
	}
}

func (table *Table) scanIndices(v interface{}) {
	tv := reflect.TypeOf(v)
	num := tv.NumField()
	for i := 0; i < num; i++ {
		table.AddIndex(CreateIndexByField(tv.Field(i)))
	}
}

func NewTable(name string, d Dialect) *Table {
	return &Table{Name: name, Dialect: d}
}

func (t *Table) createTableSQL(ifNotExists bool) string {
	var ifNotExistsStr string
	if ifNotExists {
		ifNotExistsStr = "IF NOT EXISTS "
	} else {
		ifNotExistsStr = ""
	}
	prefix := fmt.Sprintf("CREATE TABLE %s%s", ifNotExistsStr, t.Name)
	columns := make([]string, len(t.Columns))
	for i, _ := range columns {
		columns[i] = t.Columns[i].ToSQL(t.Dialect.Column())
	}
	columnDef := strings.Join(columns, ",\n")
	return fmt.Sprintf("%s(\n%s\n)%s", prefix, columnDef, t.Dialect.CreateTableSuffix())
}

func (t *Table) CreateTableSQL() string {
	return t.createTableSQL(false)
}

func (t *Table) CreateIndexSQLs() []string {
	result := make([]string, len(t.Indices))

	for i, _ := range t.Indices {
		result[i] = t.Indices[i].ToSQL(t.Name)
	}
	return result
}
func (t *Table) CreateTableIfNotExistsSQL() string {
	return t.createTableSQL(true)
}

func (t *Table) String() string {
	ddl := t.CreateTableIfNotExistsSQL()
	indices := t.CreateIndexSQLs()
	return fmt.Sprintf("%s;\n%s;\n", ddl, strings.Join(indices, ";\n"))
}

func (t *Table) AddColumn(c *Column) {
	if t.Columns == nil {
		t.Columns = make([]*Column, 0)
	}
	t.Columns = append(t.Columns, c)
}

func (t *Table) AddIndex(c *Index) {
	if t.Indices == nil {
		t.Indices = make([]*Index, 0)
	}
	t.Indices = append(t.Indices, c)
}
