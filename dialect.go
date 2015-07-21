package schemabuilder

type Dialect interface {
	Column() ColumnMapper
	CreateTableSuffix() string
	Quote(string) string
}

type ColumnMapper interface {
	DataType(v interface{}, autoIncr bool, size uint64) string
	AutoIncrement() string
	Quote(string) string
}
