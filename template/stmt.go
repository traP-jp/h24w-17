package template

import (
	"context"
	"database/sql/driver"

	"github.com/motoki317/sc"
)

var (
	stmtKey = struct{}{}
	argsKey = struct{}{}
)

// TODO: generate
// NOTE: no write happens to this map, so it's safe to use in concurrent environment
var caches = map[string]sc.Cache[string, driver.Rows]{}

type CustomCacheStatement struct {
	inner driver.Stmt
	query string
	// TODO: add intermediate format
}

func (s *CustomCacheStatement) Close() error {
	return s.inner.Close()
}

func (s *CustomCacheStatement) NumInput() int {
	return s.inner.NumInput()
}

func (s *CustomCacheStatement) Exec(args []driver.Value) (driver.Result, error) {
	return s.inner.Exec(args)
}

func (s *CustomCacheStatement) Query(args []driver.Value) (driver.Rows, error) {
	ctx := context.WithValue(context.Background(), stmtKey, s)
	ctx = context.WithValue(ctx, argsKey, args)
	return caches[cacheName(s.query)].Get(ctx, createCacheKey(s.query, args))
}

func createCacheKey(query string, args []driver.Value) string {
	// TODO: implement
	_ = query
	_ = args
	return "TODO"
}

func cacheName(query string) string {
	// TODO: implement
	return query
}

// Example of generated cache code:
//
// var selectCache = sc.NewMust(replaceFn, 0, 0)
//
// func replaceFn(ctx context.Context, key string) (*CacheRows, error) {
// 	stmt := ctx.Value(stmtKey).(*CustomCacheStatement)
// 	args := ctx.Value(argsKey).([]driver.Value)
// 	rows, err := stmt.inner.Query(args)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return NewCachedRows(rows), nil
// }
