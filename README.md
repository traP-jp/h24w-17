# isuc

![isuc](https://github.com/user-attachments/assets/4df2ac31-e131-4124-bb00-79cd6f1fb0d3)

> [!WARNING]
> Not Recommended for Production

## How to use

1. [Install isuc CLI](#install-isuc-cli)
2. Extract the queries [statically](#static-extractor) or [dynamically](#dynamic-extractor)
3. [Get your table schema](#getting-table-schemas)
4. [Generate cache plan](#generate-cache-plan)
5. [Generate the driver](#generate-the-driver)
6. [Switch the driver](#switch-the-driver)

### Install isuc CLI

```sh
go install github.com/traP-jp/isuc/cli/isuc@latest
```

### Extractor

#### Static Extractor

- `--out` represents the destination file of the extracted queries.
  - Set to `extracted.sql` by default

```sh
isuc extract --out extracted.sql /path/to/your/codebase/dir
```

#### Dynamic Extractor

1. add import statement

```go
import (
  dynamic_extractor "github.com/traP-jp/isuc/extractor/dynamic"
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

### Getting Table Schemas

If you do not have the `schema.sql`, you can generate it by the command below (you should change the auth info)

```sh
DATABASE='YOUR_DATABASE'
USER='YOUR_DATABASE_USER'
PASSWORD='YOUR_DATABASE_PASSWORD'
HOST='YOUR_DATABASE_HOST'
mysql -u "$USER" -p"$PASSWORD" -h "$HOST" -N -e "SHOW TABLES FROM $DATABASE" | while read table; do mysql -u "$USER" -p"$PASSWORD" -h "$HOST" -e "SHOW CREATE TABLE $DATABASE.\`$table\`" | awk 'NR>1 {$1=""; print substr($0,2) ";"}' | sed 's/\\n/\n/g'; done > schema.sql
```

### Generate Cache Plan

```sh
isuc analyze --sql extracted.sql --schema schema.sql --out isuc.yaml
```

- `--sql` represents extracted queries (via the static/dynamic extractor)
  - Set to `extracted.sql` by default
- `--schema` represents the table schema sql
  - Set to `schema.sql` by default
- `--out` is the destination file of the cache plan
  - Set to `isuc.yaml` by default

### Generate the driver

```sh
isuc generate --plan isuc.yaml --schema schema.sql <dist>
```

- `--plan` represents generated cache plan
  - Set to `isuc.yaml` by default
- `--schema` represents the table schema sql
  - Set to `schema.sql` by default
- `<dist>` represents the destination folder (must exist) that the generated driver will be stored into

### Switch the driver

Rewrite the section of connecting to a database.

```diff
- db, err := sql.Open("mysql", {dsn})
+ db, err := sql.Open("mysql+cache", {dsn})
```

## Appendix

### Cache Plan Format

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

type Order = {
  column: string
  order: 'asc' | 'desc'
}

type SelectQuery = CachableSelectQuery | NonCachableSelectQuery

type CachableSelectQuery = {
  type: 'select'
  query: string
  cache: true
  table: string
  targets: string[]
  conditions: Condition[]
  orders: Order[]
}

type NonCachableSelectQuery = {
  type: 'select'
  query: string
  cache: false
  table?: string
}

type UpdateQuery = {
  type: 'update'
  query: string
  table: string
  targets: string[]
  conditions: Condition[]
  orders: Order[]
}

type DeleteQuery = {
  type: 'delete'
  query: string
  table: string
  conditions: Condition[]
  orders: Order[]
}

type InsertQuery = {
  type: 'insert'
  query: string
  table: string
}
```
