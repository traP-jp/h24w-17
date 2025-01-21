package dynamic_extractor

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/traP-jp/h24w-17/normalizer"
)

func init() {
	sql.Register("mysql+analyzer", AnalyzerDriver{})
}

type AnalyzerDriver struct{}

func (d AnalyzerDriver) Open(dsn string) (driver.Conn, error) {
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
	return &AnalyzerConn{inner: conn}, nil
}

var _ driver.Driver = AnalyzerDriver{}

type AnalyzerConn struct {
	inner driver.Conn
}

func processQuery(query string) error {
	normalized, err := normalizer.NormalizeQuery(query)
	if err != nil {
		fmt.Printf("[WARN] failed to normalize query: %v\n", err)
		return err
	}
	addQuery(normalized.Query)
	return nil
}

func (c *AnalyzerConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	i, ok := c.inner.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	processQuery(query)
	return i.ExecContext(ctx, query, args)
}

func (c *AnalyzerConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	i, ok := c.inner.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	processQuery(query)
	return i.QueryContext(ctx, query, args)
}

func (c *AnalyzerConn) Prepare(query string) (driver.Stmt, error) {
	return c.inner.Prepare(query)
}

func (c *AnalyzerConn) Close() error {
	return c.inner.Close()
}

func (c *AnalyzerConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *AnalyzerConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if i, ok := c.inner.(driver.ConnBeginTx); ok {
		return i.BeginTx(ctx, opts)
	}
	return c.inner.Begin()
}

var _ driver.Conn = &AnalyzerConn{}
var _ driver.QueryerContext = &AnalyzerConn{}
var _ driver.ExecerContext = &AnalyzerConn{}
