package schemabuilder

type Dialect interface {
	Column() ColumnMapper
	CreateTableSuffix() string
}

type ColumnMapper interface {
	DataType(v interface{}, size uint64) string
	AutoIncrement() string
}
