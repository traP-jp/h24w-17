package dynamic_extractor

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"regexp"
	"strings"

	"github.com/go-sql-driver/mysql"
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

var sqlPattern = regexp.MustCompile(`(?i)\b(SELECT|INSERT|UPDATE|DELETE)\b`)
var replacePattern = regexp.MustCompile(`\s+`)

func processQuery(query string) {
	if !sqlPattern.MatchString(query) {
		return
	}
	query = strings.ReplaceAll(query, "\n", " ")
	query = replacePattern.ReplaceAllString(query, " ")
	addQuery(query)
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
