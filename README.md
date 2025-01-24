# (名称未定)

## How to use

## Extractor

### Static Extractor

TODO

### Dynamic Extractor

1. add import statement

```go
import (
  dynamic_extractor "github.com/traP-jp/h24w-17/extractor/dynamic"
)
```

2. add the code to start server

```go
func main() {
  dynamic_extractor.StartServer()
  // ...
}
```

3. replace driver `mysql` with `mysql+analyzer`
4. running your application
5. access `http://localhost:39393` and get the query list

## Getting Table Schemas

```sh
DATABASE='isupipe'
mysql -u root -ppass -h 127.0.0.1 -N -e "SHOW TABLES FROM $DATABASE" | while read table; do mysql -u root -ppass -h 127.0.0.1 -e "SHOW CREATE TABLE $DATABASE.\`$table\`" | awk 'NR>1 {$1=""; print substr($0,2) ";"}' | sed 's/\\n/\n/g'; done > schema.sql
```

## Cache Plan

### Format

```ts
type Format = {
  queries: Query[]
}

type Query = SelectQuery | UpdateQuery | DeleteQuery | InsertQuery

type Placeholder = {
  index: number;
  extra?: boolean
}

type Condition = {
  column: string
  operator: 'eq' | 'in'
  placeholder: Placeholder
}

type SelectQuery = CachableSelectQuery | NonCachableSelectQuery

type CachableSelectQuery = {
  type: 'select'
  query: string
  cache: true
  table: string
  targets: string[]
  conditions: Condition[]
  orders: {
    column: string
    order: 'asc' | 'desc'
  }[]
}

type NonCachableSelectQuery = {
  type: 'select'
  query: string
  cache: false
}

type UpdateQuery = {
  type: 'update'
  query: string
  table: string
  targets: string[]
  conditions: Condition[]
}

type DeleteQuery = {
  type: 'delete'
  query: string
  table: string
  conditions: Condition[]
}

type InsertQuery = {
  type: 'insert'
  query: string
  table: string
}
```
