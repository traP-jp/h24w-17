package cache

import (
	"context"
	"database/sql/driver"
	"log"
	"slices"

	"github.com/traP-jp/isuc/domains"
	"github.com/traP-jp/isuc/normalizer"
)

// NOTE: no write happens to this map, so it's safe to use in concurrent environment
var caches = make(map[string]*cacheWithInfo)

var cacheByTable = make(map[string][]*cacheWithInfo)

var _ driver.Stmt = &customCacheStatement{}

type customCacheStatement struct {
	inner    driver.Stmt
	conn     *cacheConn
	rawQuery string
	// query is the normalized query
	query     string
	queryInfo domains.CachePlanQuery
}

func (s *customCacheStatement) Close() error {
	return s.inner.Close()
}

func (s *customCacheStatement) NumInput() int {
	return s.inner.NumInput()
}

func (s *customCacheStatement) Exec(args []driver.Value) (driver.Result, error) {
	switch s.queryInfo.Type {
	case domains.CachePlanQueryType_INSERT:
		return s.execInsert(args)
	case domains.CachePlanQueryType_UPDATE:
		return s.execUpdate(args)
	case domains.CachePlanQueryType_DELETE:
		return s.execDelete(args)
	}

	return s.inner.(driver.StmtExecContext).ExecContext(context.Background(), valueToNamedValue(args))
}

func (s *customCacheStatement) execInsert(args []driver.Value) (driver.Result, error) {
	handleInsertQuery(s.query, *s.queryInfo.Insert, args)
	return s.inner.(driver.StmtExecContext).ExecContext(context.Background(), valueToNamedValue(args))
}

func (s *customCacheStatement) execUpdate(args []driver.Value) (driver.Result, error) {
	handleUpdateQuery(*s.queryInfo.Update, args)
	nvarsgs := valueToNamedValue(args)
	return s.inner.(driver.StmtExecContext).ExecContext(context.Background(), nvarsgs)
}

func (s *customCacheStatement) execDelete(args []driver.Value) (driver.Result, error) {
	handleDeleteQuery(*s.queryInfo.Delete, args)
	nvargs := valueToNamedValue(args)
	return s.inner.(driver.StmtExecContext).ExecContext(context.Background(), nvargs)
}

func (c *cacheConn) ExecContext(ctx context.Context, rawQuery string, nvargs []driver.NamedValue) (driver.Result, error) {
	inner, ok := c.inner.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	normalizedQuery := normalizer.NormalizeQuery(rawQuery)

	queryInfo, ok := queryMap[normalizedQuery]
	if !ok {
		log.Println("unknown query:", normalizedQuery)
		PurgeAllCaches()
		return inner.ExecContext(ctx, rawQuery, nvargs)
	}

	var res driver.Result
	var err error
	switch queryInfo.Type {
	case domains.CachePlanQueryType_INSERT:
		res, err = c.execInsert(ctx, rawQuery, queryInfo, nvargs, inner)
	case domains.CachePlanQueryType_UPDATE:
		res, err = c.execUpdate(ctx, rawQuery, queryInfo, nvargs, inner)
	case domains.CachePlanQueryType_DELETE:
		res, err = c.execDelete(ctx, rawQuery, queryInfo, nvargs, inner)
	default:
		res, err = inner.ExecContext(ctx, rawQuery, nvargs)
	}

	if !c.tx {
		c.cleanUp.do()
		c.cleanUp.reset()
	} else {
		// cleanups are deferred until the transaction is committed
		// because we need to forget the cache only if the transaction is committed

		// update the cache
		for _, c := range c.cleanUp.purge {
			c.updateTx()
		}
		for _, forget := range c.cleanUp.forget {
			forget.cache.updateByKeyTx(forget.key)
		}
	}

	return res, err
}

func (c *cacheConn) execInsert(ctx context.Context, rawQuery string, queryInfo domains.CachePlanQuery, nvargs []driver.NamedValue, inner driver.ExecerContext) (driver.Result, error) {
	args := make([]driver.Value, 0, len(nvargs))
	for _, nv := range nvargs {
		args = append(args, nv.Value)
	}

	cleanUp := handleInsertQuery(queryInfo.Query, *queryInfo.Insert, args)
	c.cleanUp.append(cleanUp)

	return inner.ExecContext(ctx, rawQuery, nvargs)
}

func (c *cacheConn) execUpdate(ctx context.Context, rawQuery string, queryInfo domains.CachePlanQuery, nvargs []driver.NamedValue, inner driver.ExecerContext) (driver.Result, error) {
	args := namedToValue(nvargs)

	cleanUp := handleUpdateQuery(*queryInfo.Update, args)
	c.cleanUp.append(cleanUp)

	return inner.ExecContext(ctx, rawQuery, nvargs)
}

func (c *cacheConn) execDelete(ctx context.Context, rawQuery string, queryInfo domains.CachePlanQuery, nvargs []driver.NamedValue, inner driver.ExecerContext) (driver.Result, error) {
	args := namedToValue(nvargs)

	cleanUp := handleDeleteQuery(*queryInfo.Delete, args)
	c.cleanUp.append(cleanUp)

	return inner.ExecContext(ctx, rawQuery, nvargs)
}

func (s *customCacheStatement) Query(args []driver.Value) (driver.Rows, error) {
	ctx := context.WithValue(context.Background(), stmtKey{}, s)
	ctx = context.WithValue(ctx, argsKey{}, args)

	conditions := s.queryInfo.Select.Conditions
	// if query is like "SELECT * FROM table WHERE cond IN (?, ?, ?, ...)"
	if len(conditions) == 1 && conditions[0].Operator == domains.CachePlanOperator_IN {
		return s.inQuery(args)
	}

	rows, err := caches[cacheName(s.query)].Get(ctx, cacheKey(args))
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (s *customCacheStatement) inQuery(args []driver.Value) (driver.Rows, error) {
	// "SELECT * FROM table WHERE cond IN (?, ?, ...)"
	// separate the query into multiple queries and merge the results
	table := s.queryInfo.Select.Table
	condIdx := s.queryInfo.Select.Conditions[0].Placeholder.Index
	condValues := args[condIdx:]

	// find the query "SELECT * FROM table WHERE cond = ?"
	var cache *cacheWithInfo
	for _, c := range cacheByTable[table] {
		if len(c.info.Conditions) == 1 && c.info.Conditions[0].Column == s.queryInfo.Select.Conditions[0].Column && c.info.Conditions[0].Operator == domains.CachePlanOperator_EQ {
			cache = c
		}
	}
	if cache == nil {
		return s.inner.(driver.StmtQueryContext).QueryContext(context.Background(), valueToNamedValue(args))
	}

	allRows := make([]*cacheRows, 0, len(condValues))
	for _, condValue := range condValues {
		// prepare new statement
		stmt, err := s.conn.Prepare(cache.query)
		if err != nil {
			return nil, err
		}
		ctx := context.WithValue(context.Background(), stmtKey{}, stmt)
		ctx = context.WithValue(ctx, argsKey{}, []driver.Value{condValue})
		rows, err := cache.Get(ctx, cacheKey([]driver.Value{condValue}))
		if err != nil {
			return nil, err
		}
		allRows = append(allRows, rows)
	}

	return mergeCachedRows(allRows), nil
}

func (c *cacheConn) QueryContext(ctx context.Context, rawQuery string, nvargs []driver.NamedValue) (driver.Rows, error) {
	inner, ok := c.inner.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	normalizedQuery := normalizer.NormalizeQuery(rawQuery)

	queryInfo, ok := queryMap[normalizedQuery]
	if !ok {
		return inner.QueryContext(ctx, rawQuery, nvargs)
	}
	if queryInfo.Type != domains.CachePlanQueryType_SELECT || !queryInfo.Select.Cache {
		return inner.QueryContext(ctx, rawQuery, nvargs)
	}

	conditions := queryInfo.Select.Conditions
	// if query is like "SELECT * FROM table WHERE cond IN (?, ?, ?, ...)"
	if len(conditions) == 1 && conditions[0].Operator == domains.CachePlanOperator_IN {
		return c.inQuery(ctx, rawQuery, nvargs, inner)
	}

	args := make([]driver.Value, len(nvargs))
	for i, nv := range nvargs {
		args[i] = nv.Value
	}

	cache := caches[queryInfo.Query]
	key := cacheKey(args)

	if c.tx && cache.isNewerThan(key, c.txStart) {
		// cache is newer than the transaction start time
		// we should not use the cache
		return inner.QueryContext(ctx, rawQuery, nvargs)
	}

	cachectx := context.WithValue(ctx, namedValueArgsKey{}, nvargs)
	cachectx = context.WithValue(cachectx, queryerCtxKey{}, inner)
	cachectx = context.WithValue(cachectx, queryKey{}, rawQuery)
	rows, err := cache.Get(cachectx, key)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (c *cacheConn) inQuery(ctx context.Context, query string, args []driver.NamedValue, inner driver.QueryerContext) (driver.Rows, error) {
	// "SELECT * FROM table WHERE cond IN (?, ?, ...)"
	// separate the query into multiple queries and merge the results
	normalizedQuery := normalizer.NormalizeQuery(query)

	queryInfo := queryMap[normalizedQuery]
	table := queryInfo.Select.Table
	condIdx := queryInfo.Select.Conditions[0].Placeholder.Index
	condValues := args[condIdx:]

	// find the query "SELECT * FROM table WHERE cond = ?"
	var cache *cacheWithInfo
	for _, c := range cacheByTable[table] {
		if len(c.info.Conditions) == 1 && c.info.Conditions[0].Column == queryInfo.Select.Conditions[0].Column && c.info.Conditions[0].Operator == domains.CachePlanOperator_EQ {
			cache = c
		}
	}
	if cache == nil {
		return inner.QueryContext(ctx, query, args)
	}

	allRows := make([]*cacheRows, 0, len(condValues))
	for _, condValue := range condValues {
		nvargs := []driver.NamedValue{condValue}
		cacheCtx := context.WithValue(ctx, queryKey{}, cache.query)
		cacheCtx = context.WithValue(cacheCtx, queryerCtxKey{}, inner)
		cacheCtx = context.WithValue(cacheCtx, namedValueArgsKey{}, nvargs)
		rows, err := cache.Get(cacheCtx, cacheKey([]driver.Value{condValue.Value}))
		if err != nil {
			return nil, err
		}
		allRows = append(allRows, rows)
	}

	return mergeCachedRows(allRows), nil
}

func handleInsertQuery(query string, queryInfo domains.CachePlanInsertQuery, insertValues []driver.Value) (cleanUp cleanUpTask) {
	table := queryInfo.Table
	insertArgs, _ := normalizer.NormalizeArgs(query)

	rows := slices.Chunk(insertValues, len(queryInfo.Columns))

	for _, cache := range cacheByTable[table] {
		if cache.uniqueOnly {
			// no need to purge
			continue
		}

		cacheConditions := cache.info.Conditions
		isComplexQuery := len(cacheConditions) != 1 || len(insertArgs.ExtraArgs) > 0 || cacheConditions[0].Operator != domains.CachePlanOperator_EQ
		if isComplexQuery {
			cleanUp.purge = append(cleanUp.purge, cache)
			continue
		}

		cacheCondition := cacheConditions[0]
		insertColumnIdx := slices.Index(queryInfo.Columns, cacheCondition.Column)
		if insertColumnIdx >= 0 {
			// insert query: "INSERT INTO table (col1, col2, ...) VALUES (?, ?, ...), (?, ?, ...), ..."
			// select query: "SELECT * FROM table WHERE col1 = ?"
			// forget the cache
			for row := range rows {
				cleanUp.forget = append(cleanUp.forget, forgetTask{cache, cacheKey([]driver.Value{row[insertColumnIdx]})})
			}
		} else {
			cleanUp.purge = append(cleanUp.purge, cache)
		}
	}

	return cleanUp
}

func handleUpdateQuery(queryInfo domains.CachePlanUpdateQuery, args []driver.Value) cleanUpTask {
	// TODO: support composite primary key and other unique key
	table := queryInfo.Table
	updateConditions := queryInfo.Conditions

	var cleanUp cleanUpTask

	// if query is NOT "UPDATE `table` SET ... WHERE `unique_col` = ?"
	if !isSingleUniqueCondition(updateConditions, table) {
		for _, cache := range cacheByTable[table] {
			if !usedBySelectQuery(cache.info.Targets, queryInfo.Targets) {
				// no need to purge because the cache does not contain the updated column
				continue
			}
			cleanUp.purge = append(cleanUp.purge, cache)
		}
		return cleanUp
	}

	updateCondition := updateConditions[0]
	uniqueValue := args[updateCondition.Placeholder.Index]

	for _, cache := range cacheByTable[table] {
		if !usedBySelectQuery(cache.info.Targets, queryInfo.Targets) {
			// no need to purge because the cache does not contain the updated column
			continue
		}

		cacheConditions := cache.info.Conditions
		if isSingleUniqueCondition(cacheConditions, table) && cacheConditions[0].Column == updateCondition.Column {
			// forget only the updated row
			cleanUp.forget = append(cleanUp.forget, forgetTask{cache, cacheKey([]driver.Value{uniqueValue})})
		} else {
			cleanUp.purge = append(cleanUp.purge, cache)
		}
	}

	return cleanUp
}

func handleDeleteQuery(queryInfo domains.CachePlanDeleteQuery, args []driver.Value) cleanUpTask {
	table := queryInfo.Table

	var cleanUp cleanUpTask

	// if query is like "DELETE FROM table WHERE unique = ?"
	var deleteByUnique bool
	if len(queryInfo.Conditions) == 1 {
		condition := queryInfo.Conditions[0]
		column := tableSchema[table].Columns[condition.Column]
		deleteByUnique = (column.IsPrimary || column.IsUnique) && condition.Operator == domains.CachePlanOperator_EQ
	}
	if !deleteByUnique {
		// we should purge all cache
		cleanUp.purge = append(cleanUp.purge, cacheByTable[table]...)
		return cleanUp
	}

	uniqueValue := args[queryInfo.Conditions[0].Placeholder.Index]

	for _, cache := range cacheByTable[table] {
		if cache.uniqueOnly {
			// query like "SELECT * FROM table WHERE pk = ?"
			// we should forget the cache
			cleanUp.forget = append(cleanUp.forget, forgetTask{cache, cacheKey([]driver.Value{uniqueValue})})
		} else {
			cleanUp.purge = append(cleanUp.purge, cache)
		}
	}

	return cleanUp
}

func usedBySelectQuery(selectTarget []string, updateTarget []domains.CachePlanUpdateTarget) bool {
	for _, target := range updateTarget {
		inSelectTarget := slices.ContainsFunc(selectTarget, func(selectTarget string) bool {
			return selectTarget == target.Column
		})
		if inSelectTarget {
			return true
		}
	}
	return false
}

func isSingleUniqueCondition(conditions []domains.CachePlanCondition, table string) bool {
	if len(conditions) != 1 {
		return false
	}
	condition := conditions[0]
	column := tableSchema[table].Columns[condition.Column]
	return (column.IsPrimary || column.IsUnique) && condition.Operator == domains.CachePlanOperator_EQ
}

func valueToNamedValue(args []driver.Value) []driver.NamedValue {
	nvargs := make([]driver.NamedValue, len(args))
	for i, arg := range args {
		nvargs[i] = driver.NamedValue{Ordinal: i + 1, Value: arg}
	}
	return nvargs
}

func namedToValue(nvargs []driver.NamedValue) []driver.Value {
	args := make([]driver.Value, len(nvargs))
	for i, nv := range nvargs {
		args[i] = nv.Value
	}
	return args
}

type forgetTask struct {
	cache *cacheWithInfo
	key   string
}

type cleanUpTask struct {
	purge  []*cacheWithInfo
	forget []forgetTask
}

func (c *cleanUpTask) reset() {
	c.purge = c.purge[:0]
	c.forget = c.forget[:0]
}

func (c *cleanUpTask) do() {
	for _, cache := range c.purge {
		cache.Purge()
	}
	for _, forget := range c.forget {
		forget.cache.Forget(forget.key)
	}
}

func (c *cleanUpTask) append(tasks cleanUpTask) {
	c.purge = append(c.purge, tasks.purge...)
	c.forget = append(c.forget, tasks.forget...)
}
