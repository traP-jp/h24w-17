package template

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/motoki317/sc"
	"github.com/traP-jp/h24w-17/domains"
	"github.com/traP-jp/h24w-17/normalizer"
)

type (
	stmtKey       struct{}
	argsKey       struct{}
	queryerCtxKey struct{}
	namedArgsKey  struct{}
)

type cacheWithInfo struct {
	info  domains.CachePlanSelectQuery
	cache *sc.Cache[string, *CacheRows]
}

// NOTE: no write happens to this map, so it's safe to use in concurrent environment
var caches = make(map[string]cacheWithInfo)

var cacheByTable = make(map[string][]*sc.Cache[string, *CacheRows])

func ExportMetrics() string {
	res := ""
	for query, cache := range caches {
		res += "query: " + query + "\n"
		res += cache.cache.Stats().String() + "\n"
	}
	return res
}

type CustomCacheStatement struct {
	inner    driver.Stmt
	rawQuery string
	// query is the normalized query
	query           string
	extraConditions []domains.CachePlanCondition
	queryInfo       domains.CachePlanQuery
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
	ctx := context.WithValue(context.Background(), stmtKey{}, s)
	ctx = context.WithValue(ctx, argsKey{}, args)
	rows, err := caches[cacheName(s.query)].cache.Get(ctx, cacheKey(args))
	if err != nil {
		return nil, err
	}
	rows.mu.Lock()
	defer rows.mu.Unlock()
	if rows.cached {
		return rows.Clone(), nil
	}

	return rows, nil
}

func (c *CacheConn) QueryContext(ctx context.Context, rawQuery string, nvargs []driver.NamedValue) (driver.Rows, error) {
	normalized, err := normalizer.NormalizeQuery(rawQuery)
	if err != nil {
		return nil, err
	}

	inner, ok := c.inner.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	queryInfo, ok := queryMap[normalized.Query]
	if !ok {
		return inner.QueryContext(ctx, rawQuery, nvargs)
	}
	if queryInfo.Type != domains.CachePlanQueryType_SELECT || !queryInfo.Select.Cache {
		return inner.QueryContext(ctx, rawQuery, nvargs)
	}

	args := make([]driver.Value, len(nvargs))
	for i, nv := range nvargs {
		args[i] = nv.Value
	}

	cache := caches[queryInfo.Query].cache
	cacheKey := cacheKey(args)
	cachectx := context.WithValue(ctx, namedArgsKey{}, nvargs)
	cachectx = context.WithValue(cachectx, queryerCtxKey{}, inner)
	rows, err := cache.Get(cachectx, cacheKey)
	if err != nil {
		return nil, err
	}

	rows.mu.Lock()
	defer rows.mu.Unlock()
	if rows.cached {
		return rows.Clone(), nil
	}

	return rows, nil
}

func cacheName(query string) string {
	return query
}

func cacheKey(args []driver.Value) string {
	var b strings.Builder
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			b.WriteString(v)
		case []byte:
			b.Write(v)
		default:
			fmt.Fprintf(&b, "%v", v)
		}
		// delimiter
		b.WriteByte(0)
	}
	return b.String()
}

func replaceFn(ctx context.Context, key string) (*CacheRows, error) {
	queryerCtx, ok := ctx.Value(queryerCtxKey{}).(driver.QueryerContext)
	if ok {
		nvargs := ctx.Value(namedArgsKey{}).([]driver.NamedValue)
		rows, err := queryerCtx.QueryContext(ctx, key, nvargs)
		if err != nil {
			return nil, err
		}
		return NewCachedRows(rows), nil
	}

	stmt := ctx.Value(stmtKey{}).(*CustomCacheStatement)

	args := ctx.Value(argsKey{}).([]driver.Value)
	rows, err := stmt.inner.Query(args)
	if err != nil {
		return nil, err
	}

	return NewCachedRows(rows), nil
}
