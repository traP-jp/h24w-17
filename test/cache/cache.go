package cache

import (
	"context"
	"database/sql/driver"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/motoki317/sc"
	"github.com/traP-jp/isuc/domains"
)

type cacheWithInfo struct {
	*sc.Cache[string, *cacheRows]
	query           string
	info            domains.CachePlanSelectQuery
	uniqueOnly      bool         // if true, query is like "SELECT * FROM table WHERE pk = ?"
	lastUpdate      atomic.Int64 // time.Time.UnixNano()
	lastUpdateByKey syncMap[int64]
	replaceTime     atomic.Int64
}

func (c *cacheWithInfo) updateTx() {
	c.lastUpdate.Store(time.Now().UnixNano())
}

func (c *cacheWithInfo) updateByKeyTx(key string) {
	c.lastUpdateByKey.Store(key, time.Now().UnixNano())
}

func (c *cacheWithInfo) isNewerThan(key string, t int64) bool {
	if c.lastUpdate.Load() > t {
		return true
	}
	if update, ok := c.lastUpdateByKey.Load(key); ok && update > t {
		return true
	}
	return false
}

func (c *cacheWithInfo) RecordReplaceTime(time time.Duration) {
	c.replaceTime.Add(time.Nanoseconds())
}

type (
	queryKey          struct{}
	stmtKey           struct{}
	argsKey           struct{}
	queryerCtxKey     struct{}
	namedValueArgsKey struct{}
	cacheWithInfoKey  struct{}
)

func ExportMetrics() string {
	cacheList := make([]*cacheWithInfo, 0, len(caches))
	for _, cache := range caches {
		cacheList = append(cacheList, cache)
	}
	sort.SliceStable(cacheList, func(i, j int) bool {
		return cacheList[i].replaceTime.Load() < cacheList[j].replaceTime.Load()
	})
	res := ""
	for _, cache := range cacheList {
		stats := cache.Stats()
		progress := "["
		for i := 0; i < 20; i++ {
			if i < int(stats.HitRatio()*20) {
				progress += "#"
			} else {
				progress += "-"
			}
		}
		progress += "]"
		replaceTime := time.Duration(cache.replaceTime.Load()).String()
		statsStr := fmt.Sprintf("%s (%.2f%% - %d/%d)\n%d replace (%s) / size = %d", progress, stats.HitRatio()*100, stats.Hits, stats.Misses+stats.Hits, stats.Replacements, replaceTime, stats.Size)
		res += fmt.Sprintf("query: \"%s\"\n%s\n\n", cache.query, statsStr)
	}
	return res
}

type CacheStats struct {
	Query    string
	HitRatio float64
	Hits     int
	Misses   int
}

func ExportCacheStats() map[string]CacheStats {
	res := make(map[string]CacheStats)
	for query, cache := range caches {
		stats := cache.Stats()
		res[query] = CacheStats{
			Query:    query,
			HitRatio: stats.HitRatio(),
			Hits:     int(stats.Hits),
			Misses:   int(stats.Misses),
		}
	}
	return res
}

func PurgeAllCaches() {
	for _, cache := range caches {
		cache.Purge()
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
	cache := ctx.Value(cacheWithInfoKey{}).(*cacheWithInfo)
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		cache.RecordReplaceTime(elapsed)
	}()

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

type syncMap[T any] struct {
	m sync.Map
}

func (m *syncMap[T]) Load(key string) (T, bool) {
	var zero T
	v, ok := m.m.Load(key)
	if !ok {
		return zero, false
	}
	return v.(T), true
}

func (m *syncMap[T]) Store(key string, value T) {
	m.m.Store(key, value)
}
