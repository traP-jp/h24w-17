package cache

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/motoki317/sc"
	"github.com/traP-jp/isuc/domains"
	"github.com/traP-jp/isuc/normalizer"
)

var queryMap = make(map[string]domains.CachePlanQuery)

var tableSchema = make(map[string]domains.TableSchema)

const cachePlanRaw = `queries:
  - query: SELECT * FROM ` + "`" + `users` + "`" + ` WHERE ` + "`" + `id` + "`" + ` = ?;
    type: select
    table: users
    cache: true
    targets:
      - id
      - name
      - age
      - group_id
      - created_at
    conditions:
      - column: id
        operator: eq
        placeholder:
          index: 0
  - query: SELECT * FROM ` + "`" + `users` + "`" + ` WHERE ` + "`" + `id` + "`" + ` IN (?);
    type: select
    table: users
    cache: true
    targets:
      - id
      - name
      - age
      - group_id
      - created_at
    conditions:
      - column: id
        operator: in
        placeholder:
          index: 0
  - query: SELECT * FROM ` + "`" + `users` + "`" + ` WHERE ` + "`" + `group_id` + "`" + ` = ?;
    type: select
    table: users
    cache: true
    targets:
      - id
      - name
      - age
      - group_id
      - created_at
    conditions:
      - column: group_id
        operator: eq
        placeholder:
          index: 0
  - query: UPDATE ` + "`" + `users` + "`" + ` SET ` + "`" + `name` + "`" + ` = ? WHERE ` + "`" + `id` + "`" + ` = ?;
    type: update
    table: users
    targets:
      - column: name
        placeholder:
          index: 0
    conditions:
      - column: id
        operator: eq
        placeholder:
          index: 1
  - query: INSERT INTO ` + "`" + `users` + "`" + ` (` + "`" + `name` + "`" + `, ` + "`" + `age` + "`" + `, ` + "`" + `created_at` + "`" + `) VALUES (?, ?, ?);
    type: insert
    table: users
    columns:
      - name
      - age
      - created_at
  - query: INSERT INTO ` + "`" + `users` + "`" + ` (` + "`" + `name` + "`" + `, ` + "`" + `age` + "`" + `, ` + "`" + `group_id` + "`" + `, ` + "`" + `created_at` + "`" + `) VALUES (?, ?, ?, ?);
    type: insert
    table: users
    columns:
      - name
      - age
      - group_id
      - created_at
`

const schemaRaw = `CREATE TABLE ` + "`" + `users` + "`" + ` (
    ` + "`" + `id` + "`" + ` INT NOT NULL AUTO_INCREMENT,
    ` + "`" + `name` + "`" + ` VARCHAR(255) NOT NULL,
    ` + "`" + `age` + "`" + ` INT NOT NULL,
    ` + "`" + `group_id` + "`" + ` INT,
    ` + "`" + `created_at` + "`" + ` DATETIME NOT NULL,
    PRIMARY KEY (` + "`" + `id` + "`" + `)
);
`

func init() {
	sql.Register("mysql+cache", CacheDriver{})

	schema, err := domains.LoadTableSchema(schemaRaw)
	if err != nil {
		panic(err)
	}
	for _, table := range schema {
		tableSchema[table.TableName] = table
	}

	plan, err := domains.LoadCachePlan(strings.NewReader(cachePlanRaw))
	if err != nil {
		panic(err)
	}

	for _, query := range plan.Queries {
		normalized := normalizer.NormalizeQuery(query.Query)
		query.Query = normalized // make sure to use normalized query
		queryMap[normalized] = *query
		if query.Type != domains.CachePlanQueryType_SELECT || !query.Select.Cache {
			continue
		}

		conditions := query.Select.Conditions
		if isSingleUniqueCondition(conditions, query.Select.Table) {
			caches[normalized] = cacheWithInfo{
				query:      normalized,
				info:       *query.Select,
				cache:      sc.NewMust(replaceFn, 10*time.Minute, 10*time.Minute),
				uniqueOnly: true,
			}
			continue
		}
		caches[query.Query] = cacheWithInfo{
			query:      query.Query,
			info:       *query.Select,
			cache:      sc.NewMust(replaceFn, 10*time.Minute, 10*time.Minute),
			uniqueOnly: false,
		}

		// TODO: if query is like "SELECT * FROM WHERE pk IN (?, ?, ...)", generate cache with query "SELECT * FROM table WHERE pk = ?"
	}

	for _, cache := range caches {
		cacheByTable[cache.info.Table] = append(cacheByTable[cache.info.Table], cache)
	}
}

var _ driver.Driver = CacheDriver{}

type CacheDriver struct{}

func (d CacheDriver) Open(dsn string) (driver.Conn, error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	c, err := mysql.NewConnector(cfg)
	if err != nil {
		return nil, err
	}
	conn, err := c.Connect(context.Background())
	if err != nil {
		return nil, err
	}
	return &cacheConn{inner: conn}, nil
}

var (
	_ driver.Conn           = &cacheConn{}
	_ driver.ConnBeginTx    = &cacheConn{}
	_ driver.Pinger         = &cacheConn{}
	_ driver.QueryerContext = &cacheConn{}
	_ driver.ExecerContext  = &cacheConn{}
)

type cacheConn struct {
	inner driver.Conn
}

func (c *cacheConn) Prepare(rawQuery string) (driver.Stmt, error) {
	normalizedQuery := normalizer.NormalizeQuery(rawQuery)

	queryInfo, ok := queryMap[normalizedQuery]
	if !ok {
		// unknown (insert, update, delete) query
		if !strings.HasPrefix(strings.ToUpper(normalizedQuery), "SELECT") {
			log.Println("unknown query:", normalizedQuery)
			PurgeAllCaches()
		}
		return c.inner.Prepare(rawQuery)
	}

	if queryInfo.Type == domains.CachePlanQueryType_SELECT && !queryInfo.Select.Cache {
		return c.inner.Prepare(rawQuery)
	}

	innerStmt, err := c.inner.Prepare(rawQuery)
	if err != nil {
		return nil, err
	}
	return &customCacheStatement{
		inner:     innerStmt,
		conn:      c,
		rawQuery:  rawQuery,
		query:     normalizedQuery,
		queryInfo: queryInfo,
	}, nil
}

func (c *cacheConn) Close() error {
	return c.inner.Close()
}

func (c *cacheConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *cacheConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if i, ok := c.inner.(driver.ConnBeginTx); ok {
		return i.BeginTx(ctx, opts)
	}
	return c.inner.Begin()
}

func (c *cacheConn) Ping(ctx context.Context) error {
	if i, ok := c.inner.(driver.Pinger); ok {
		return i.Ping(ctx)
	}
	return nil
}

var _ driver.Rows = &cacheRows{}

type cacheRows struct {
	cached  bool
	columns []string
	rows    sliceRows
}

func newCacheRows(inner driver.Rows) (*cacheRows, error) {
	r := new(cacheRows)

	err := r.cacheInnerRows(inner)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *cacheRows) clone() *cacheRows {
	if !r.cached {
		panic("cannot clone uncached rows")
	}
	return &cacheRows{
		cached:  r.cached,
		columns: r.columns,
		rows:    r.rows.clone(),
	}
}

func (r *cacheRows) Columns() []string {
	if !r.cached {
		panic("cannot get columns of uncached rows")
	}
	return r.columns
}

func (r *cacheRows) Close() error {
	r.rows.reset()
	return nil
}

func (r *cacheRows) Next(dest []driver.Value) error {
	if !r.cached {
		return fmt.Errorf("cannot get next row of uncached rows")
	}
	return r.rows.next(dest)
}

func mergeCachedRows(rows []*cacheRows) *cacheRows {
	if len(rows) == 0 {
		return nil
	}
	if len(rows) == 1 {
		return rows[0]
	}

	mergedSlice := sliceRows{}
	for _, r := range rows {
		mergedSlice.concat(r.rows)
	}

	return &cacheRows{
		cached:  true,
		columns: rows[0].columns,
		rows:    mergedSlice,
	}
}

func (r *cacheRows) cacheInnerRows(inner driver.Rows) error {
	columns := inner.Columns()
	r.columns = columns
	dest := make([]driver.Value, len(columns))

	for {
		err := inner.Next(dest)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		cachedRow := make(row, len(dest))
		for i := 0; i < len(dest); i++ {
			switch v := dest[i].(type) {
			case int64, uint64, float64, string, bool, time.Time, nil: // no need to copy
				cachedRow[i] = v
			case []byte: // copy to prevent mutation
				data := make([]byte, len(v))
				copy(data, v)
				cachedRow[i] = data
			default:
				// TODO: handle other types
				// Should we mark this row as uncacheable?
			}
		}
		r.rows.append(cachedRow)
	}

	r.cached = true

	return nil
}

type row = []driver.Value

type sliceRows struct {
	rows []row
	idx  int
}

func (r sliceRows) clone() sliceRows {
	rows := make([]row, len(r.rows))
	copy(rows, r.rows)
	return sliceRows{rows: rows}
}

func (r *sliceRows) append(row ...row) {
	r.rows = append(r.rows, row...)
}

func (r *sliceRows) concat(rows sliceRows) {
	r.rows = append(r.rows, rows.rows...)
}

func (r *sliceRows) reset() {
	r.idx = 0
}

func (r *sliceRows) next(dest []driver.Value) error {
	if r.idx >= len(r.rows) {
		r.reset()
		return io.EOF
	}
	row := r.rows[r.idx]
	r.idx++
	copy(dest, row)
	return nil
}
