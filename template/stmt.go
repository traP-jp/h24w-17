package template

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/motoki317/sc"
	"github.com/traP-jp/h24w-17/domains"
)

var (
	stmtKey = struct{}{}
	argsKey = struct{}{}
)

// TODO: generate
// NOTE: no write happens to this map, so it's safe to use in concurrent environment
var caches map[string]sc.Cache[string, driver.Rows]

var cacheByTable map[string][]sc.Cache[string, driver.Rows]

type CustomCacheStatement struct {
	inner     driver.Stmt
	query     string
	queryInfo domains.CachePlanQuery
}

func (s *CustomCacheStatement) Close() error {
	return s.inner.Close()
}

func (s *CustomCacheStatement) NumInput() int {
	return s.inner.NumInput()
}

func (s *CustomCacheStatement) Exec(args []driver.Value) (driver.Result, error) {
	switch s.queryInfo.Type {
	case domains.CachePlanQueryType_INSERT:
		return s.execInsert(args)
	case domains.CachePlanQueryType_UPDATE:
		return s.execUpdate(args)
	case domains.CachePlanQueryType_DELETE:
		return s.execDelete(args)
	}
	return s.inner.Exec(args)
}

func (s *CustomCacheStatement) execInsert(args []driver.Value) (driver.Result, error) {
	table := s.queryInfo.Insert.Table
	for _, cache := range cacheByTable[table] {
		cache.Purge()
	}
	return s.inner.Exec(args)
}

func (s *CustomCacheStatement) execUpdate(args []driver.Value) (driver.Result, error) {
	// TODO: purge only necessary cache
	table := s.queryInfo.Update.Table
	for _, cache := range cacheByTable[table] {
		cache.Purge()
	}
	return s.inner.Exec(args)
}

func (s *CustomCacheStatement) execDelete(args []driver.Value) (driver.Result, error) {
	table := s.queryInfo.Delete.Table
	for _, cache := range cacheByTable[table] {
		cache.Purge()
	}
	return s.inner.Exec(args)
}

func (s *CustomCacheStatement) Query(args []driver.Value) (driver.Rows, error) {
	ctx := context.WithValue(context.Background(), stmtKey, s)
	ctx = context.WithValue(ctx, argsKey, args)
	return caches[cacheName(s.query)].Get(ctx, createCacheKey(s.query, args))
}

func createCacheKey(query string, args []driver.Value) string {
	key := query
	for _, arg := range args {
		key += fmt.Sprintf(":%v", arg)
	}
	return key
}

func cacheName(query string) string {
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
