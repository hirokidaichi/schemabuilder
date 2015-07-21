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
	Table           *Table
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

func (c *Column) String() string {
	return c.ToSQL()
}

func (this *Column) Equals(that *Column) bool {
	return (this.ToSQL() == that.ToSQL())
}

func (c *Column) AlterAddSQL() string {
	return fmt.Sprintf("ADD %s", c.ToSQL())
}

func (c *Column) AlterDropSQL() string {
	table := c.Table
	dialect := table.Dialect
	return fmt.Sprintf("DROP %s", dialect.Quote(c.Name))
}

func (c *Column) AlterModifySQL(to *Column) string {
	return fmt.Sprintf("MODIFY %s", c.ToSQL())
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

func (c *Column) ToSQL() string {
	d := c.Table.Dialect.Column()
	return fmt.Sprintf("%s %s %s", d.Quote(c.Name), c.DataType(), c.Constraints())
}

func (c *Column) DataType() string {
	d := c.Table.Dialect.Column()
	return d.DataType(c.Type, c.IsAutoIncrement, c.TypeSize)
}

func (c *Column) Constraints() string {
	d := c.Table.Dialect.Column()
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
	Table       *Table
}

func (i *Index) ToSQL() string {
	var format string
	tableName := i.Table.Name
	d := i.Table.Dialect
	if i.IsUnique {
		format = "CREATE UNIQUE INDEX %s ON %s (%s)"
	} else {
		format = "CREATE INDEX %s ON %s (%s)"
	}
	return fmt.Sprintf(format, d.Quote(i.Name), d.Quote(tableName), strings.Join(quoteAll(d, i.ColumnNames), ","))
}

func quoteAll(d Dialect, ss []string) []string {
	quoted := make([]string, len(ss))
	for i, v := range ss {
		quoted[i] = d.Quote(v)
	}
	return quoted
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
	Name      string
	Dialect   Dialect
	Columns   []*Column
	Indices   []*Index
	Version   string
	Histories []*Table
	columnMap map[string]*Column
	indexMap  map[string]*Index
}

func tableName(structName string) (string, string) {
	splitted := strings.Split(structName, "_")
	tableName := inflect.Tableize(splitted[0])
	versionName := strings.Join(splitted[1:], "_")
	return tableName, versionName
}

func CreateTableByStruct(tableStruct interface{}, indexStruct interface{}, d Dialect) *Table {
	tableName, versionName := tableName(reflect.TypeOf(tableStruct).Name())
	table := NewTable(tableName, d)
	table.Version = versionName
	table.scanColumns(tableStruct)
	if indexStruct != nil {
		table.scanIndices(indexStruct)
	}
	return table
}

func (table *Table) scanColumns(v interface{}) {
	tv := reflect.TypeOf(v)
	vv := reflect.ValueOf(v)
	num := tv.NumField()
	for i := 0; i < num; i++ {
		field := tv.Field(i)
		inf := vv.Field(i).Interface()
		if field.Anonymous {
			table.scanColumns(inf)
			continue
		}
		table.AddColumn(CreateColumnByField(inf, field))
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
	prefix := fmt.Sprintf("CREATE TABLE %s%s", ifNotExistsStr, t.Dialect.Quote(t.Name))
	columns := make([]string, len(t.Columns))
	for i, _ := range columns {
		columns[i] = t.Columns[i].ToSQL()
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
		result[i] = t.Indices[i].ToSQL()
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

func (t *Table) GetColumn(name string) *Column {
	return t.columnMap[name]
}

func (t *Table) GetIndex(name string) *Index {
	return t.indexMap[name]
}

func (t *Table) AddColumn(c *Column) *Table {
	if t.Columns == nil {
		t.Columns = make([]*Column, 0)
	}
	if t.columnMap == nil {
		t.columnMap = make(map[string]*Column)
	}
	c.Table = t
	t.columnMap[c.Name] = c
	t.Columns = append(t.Columns, c)
	return t
}

func (t *Table) AddIndex(c *Index) *Table {
	if t.Indices == nil {
		t.Indices = make([]*Index, 0)
	}
	if t.indexMap == nil {
		t.indexMap = make(map[string]*Index)
	}
	c.Table = t
	t.indexMap[c.Name] = c
	t.Indices = append(t.Indices, c)
	return t
}

func (t *Table) AddHistory(tableStruct, indexStruct interface{}) *Table {
	history := CreateTableByStruct(tableStruct, indexStruct, t.Dialect)
	if history.Name != t.Name {
		panic(fmt.Sprintf("cannot add history %s to %s", history.Name, t.Name))
	}
	return t.AddHistoryTable(history)
}

func (t *Table) AddHistoryTable(c *Table) *Table {
	if t.Histories == nil {
		t.Histories = make([]*Table, 0)
	}
	t.Histories = append(t.Histories, c)
	return t
}

func (t *Table) lookupVersion(version string) *Table {
	if version == "current" {
		return t
	}
	for _, v := range t.Histories {
		if v.Version == version {
			return v
		}
	}
	return nil
}

func (t *Table) MigrateSQL(from, to string) (string, error) {
	result := make([]string, 0)
	fromTable := t.lookupVersion(from)
	toTable := t.lookupVersion(to)
	if fromTable == nil {
		return "", fmt.Errorf("from table not found version (%s)", from)
	}
	if toTable == nil {
		return "", fmt.Errorf("to table not found version (%s)", to)
	}
	for _, fromColumn := range fromTable.Columns {
		toColumn := toTable.GetColumn(fromColumn.Name)
		if toColumn == nil {
			result = append(result, fromColumn.AlterDropSQL())
			continue
		}
		if !fromColumn.Equals(toColumn) {
			result = append(result, fromColumn.AlterModifySQL(toColumn))
			continue
		}
	}
	for _, toColumn := range toTable.Columns {
		fromColumn := fromTable.GetColumn(toColumn.Name)
		if fromColumn == nil {
			result = append(result, toColumn.AlterAddSQL())
		}
	}
	sql := fmt.Sprintf("ALTER TABLE %s\n%s",
		t.Dialect.Quote(t.Name), strings.Join(result, ",\n"))

	return sql, nil
}

type Builder struct {
	Dialect Dialect
}

func For(d Dialect) *Builder {
	return &Builder{Dialect: d}
}

func (b *Builder) DefineTable(tableStruct, indexStruct interface{}) *Table {
	return CreateTableByStruct(tableStruct, indexStruct, b.Dialect)
}
