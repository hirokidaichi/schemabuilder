
schemabuilder - a simple sql schema builder for golang
=====

I want not O/R Mapper but a simple DDL builder. 

## Feature

+ Get a SQL as string
+ Get a Migration SQL as string 
+ MySQL and SQLite3 Supported
+ Composed index supported


## Examples

### Scanning Struct

```
package main

import (
	"github.com/hirokidaichi/schemabuilder"
	"time"
)

type Person struct {
	Id        uint64 `pk:"true",autoincrement:"true"`
	Name      string `size:"200",unique:"true"`
	Info      *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PersonIndex struct {
	key1 schemabuilder.IComposedKeys `columns:"CreatedAt,UpdatedAt"`
	key2 schemabuilder.IUniqueKeys   `columns:"CreatedAt,UpdatedAt"`
}

var builder = schemabuilder.For(schemabuilder.NewMySQLDialect("utf8", "InnoDB"))
var personSchema = builder.DefineTable(Person{}, PersonIndex{})

func main() {
	println(personSchema.String())
}

/*

CREATE TABLE IF NOT EXISTS `people`(
    `id` BIGINT NOT NULL PRIMARY KEY,
    `name` VARCHAR(200) NOT NULL,
    `info` VARCHAR(255) ,
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8;
CREATE INDEX `key1` ON `people` (`created_at`,`updated_at`);
CREATE UNIQUE INDEX `key2` ON `people` (`created_at`,`updated_at`);

*/

```

### Migration

```
package main

import (
	"fmt"
	"github.com/hirokidaichi/schemabuilder"
	"time"
)

type Person_old struct {
	Id        uint64 `pk:"true",autoincrement:"true"`
	Name      string `size:"200",unique:"true"`
	Info      *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Person_new struct {
	Person_old
	Json  *string `size:"2048"`
	Score *uint8
}

type Person Person_new

type PersonIndex struct {
	key1 schemabuilder.IComposedKeys `columns:"CreatedAt,UpdatedAt"`
	key2 schemabuilder.IUniqueKeys   `columns:"CreatedAt,UpdatedAt"`
}

var builder = schemabuilder.For(schemabuilder.NewMySQLDialect("utf8", "InnoDB"))
var personSchema = builder.DefineTable(Person{}, PersonIndex{}).
	AddHistory(Person_old{}, PersonIndex{}).
	AddHistory(Person_new{}, PersonIndex{})

func main() {
	fmt.Println(personSchema.MigrateSQL("old", "new"))
	// ALTER TABLE `people` ADD `json` VARCHAR(2048) , ADD `score` SMALLINT
}
```

## SEE ALSO

+ gorp
+ genmai
+ and other o/r mappers

## License

MIT

## Author

hirokidaichi [at] gmail.com

