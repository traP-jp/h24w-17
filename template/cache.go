package template

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strings"
)

type (
	queryKey          struct{}
	stmtKey           struct{}
	argsKey           struct{}
	queryerCtxKey     struct{}
	namedValueArgsKey struct{}
)

func ExportMetrics() string {
	res := ""
	for query, cache := range caches {
		stats := cache.cache.Stats()
		progress := "["
		for i := 0; i < 20; i++ {
			if i < int(stats.HitRatio()*20) {
				progress += "#"
			} else {
				progress += "-"
			}
		}
		statsStr := fmt.Sprintf("%s (%.2f%% - %d/%d) (%d replace) (size %d)", progress, stats.HitRatio()*100, stats.Hits, stats.Misses+stats.Hits, stats.Replacements, stats.Size)
		res += fmt.Sprintf("query: \"%s\"\n%s\n\n", query, statsStr)
	}
	return res
}

func PurgeAllCaches() {
	for _, cache := range caches {
		cache.cache.Purge()
	}
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

func replaceFn(ctx context.Context, key string) (*cacheRows, error) {
	queryerCtx, ok := ctx.Value(queryerCtxKey{}).(driver.QueryerContext)
	if ok {
		query := ctx.Value(queryKey{}).(string)
		nvargs := ctx.Value(namedValueArgsKey{}).([]driver.NamedValue)
		rows, err := queryerCtx.QueryContext(ctx, query, nvargs)
		if err != nil {
			return nil, err
		}
		cacheRows, err := newCacheRows(rows)
		if err != nil {
			return nil, err
		}
		return cacheRows.clone(), nil
	}

	stmt := ctx.Value(stmtKey{}).(*customCacheStatement)
	args := ctx.Value(argsKey{}).([]driver.Value)
	rows, err := stmt.inner.Query(args)
	if err != nil {
		return nil, err
	}
	cacheRows, err := newCacheRows(rows)
	if err != nil {
		return nil, err
	}
	return cacheRows.clone(), nil
}
