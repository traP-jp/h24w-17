# (名称未定)

## How to use

## Extractor

### Static Extractor

TODO

### Dynamic Extractor

1. add import statement

```go
import (
  _ "github.com/traP-jp/h24w-17/extractor/dynamic"
)
```

2. replace driver `mysql` with `mysql+analyzer`
3. running your application
4. access `http://localhost:39393` and get the query list

## Cache Plan

### Format

```ts
type Format = {
  queries: Query[]
}

type Query = SelectQuery | UpdateQuery | DeleteQuery | InsertQuery

type Condition = {
  column: string
  value?: number | string
}

type SelectQuery = CachableSelectQuery | NonCachableSelectQuery

type CachableSelectQuery = {
  type: 'select'
  query: string
  cache: true
  table: string
  targets: string[]
  // ?の位置と一致するような順番
  // 固定値でWHEREしている場合は最後に
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
