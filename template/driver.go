package template

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"time"

	"github.com/go-sql-driver/mysql"
)

func init() {
	sql.Register("mysql+cache", CacheDriver{})
}

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
	return &CacheConn{inner: conn}, nil
}

type CacheConn struct {
	inner driver.Conn
}

func (c *CacheConn) Prepare(query string) (driver.Stmt, error) {
	cachable := true // TODO: check if the query is cacheable from intermediate format
	if !cachable {
		return c.inner.Prepare(query)
	}

	innerStmt, err := c.inner.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &CustomCacheStatement{inner: innerStmt, query: query}, nil
}

func (c *CacheConn) Close() error {
	return c.inner.Close()
}

func (c *CacheConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *CacheConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if c, ok := c.inner.(driver.ConnBeginTx); ok {
		return c.BeginTx(ctx, opts)
	}
	return c.Begin()
}

type CacheRows struct {
	inner   driver.Rows
	cached  bool
	columns []string
	rows    sliceRows
}

func NewCachedRows(inner driver.Rows) *CacheRows {
	return &CacheRows{inner: inner}
}

type row = []driver.Value

// TODO: goroutine safe
type sliceRows struct {
	rows []row
	idx  int
}

func (r *sliceRows) append(row row) {
	r.rows = append(r.rows, row)
}

func (r *sliceRows) reset() {
	r.idx = 0
}

func (r *sliceRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.rows) {
		r.idx = 0
		return io.EOF
	}
	row := r.rows[r.idx]
	r.idx++
	copy(dest, row)
	return nil
}

func (r *CacheRows) Columns() []string {
	if r.cached {
		return r.columns
	}
	columns := r.inner.Columns()
	r.columns = make([]string, len(columns))
	copy(r.columns, columns)
	return columns
}

func (r *CacheRows) Close() error {
	if r.cached {
		r.rows.reset()
		return nil
	}
	return r.inner.Close()
}

func (r *CacheRows) Next(dest []driver.Value) error {
	if r.cached {
		return r.rows.Next(dest)
	}

	err := r.inner.Next(dest)
	if err != nil {
		if err == io.EOF {
			r.cached = true
			return err
		}
		return err
	}

	cachedRow := make(row, len(dest))
	for i := 0; i < len(dest); i++ {
		switch v := dest[i].(type) {
		case int64, float64, string, bool, time.Time, nil: // no need to copy
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

	return nil
}
