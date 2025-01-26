package analyzer

import (
	"fmt"

	"github.com/traP-jp/isuc/domains"
	"github.com/traP-jp/isuc/normalizer"
	"github.com/traP-jp/isuc/sql_parser"
)

func AnalyzeQueries(queries []string, schemas []domains.TableSchema) (domains.CachePlan, error) {
	a := newAnalyzer(schemas)
	return a.analyzeQueries(queries)
}

type analyzerError struct {
	errors []error
}

func (e analyzerError) Error() string {
	msg := ""
	for _, err := range e.errors {
		msg = fmt.Sprintf("%s\n%s", msg, err)
	}
	return msg
}

func (e analyzerError) wrap() error {
	if len(e.errors) == 0 {
		return nil
	}
	return e
}

type analyzer struct {
	schemas []domains.TableSchema
}

func newAnalyzer(schemas []domains.TableSchema) *analyzer {
	return &analyzer{schemas}
}

func (a *analyzer) analyzeQueries(queries []string) (domains.CachePlan, error) {
	plan := domains.CachePlan{}
	planErr := &analyzerError{}
	for _, query := range queries {
		week := false
		query = normalizer.NormalizeQuery(query)
		parsed, err := sql_parser.ParseSQL(query)
		if err != nil {
			weekParsed, weekErr := sql_parser.ParseSQLWeekly(query)
			if weekErr != nil {
				planErr.errors = append(planErr.errors, fmt.Errorf("failed to parse query:\nmain -> %s\nweek -> %s", err, weekErr))
				continue
			}
			parsed = weekParsed
			week = true
		}
		analyzed, err := a.analyzeQuery(parsed)
		if err != nil {
			planErr.errors = append(planErr.errors, fmt.Errorf("failed to analyze query: %s", err))
			continue
		}
		if week {
			analyzed.Query = query
		}
		if !week {
			err = a.normalizeArgs(query, &analyzed)
			if err != nil {
				planErr.errors = append(planErr.errors, fmt.Errorf("failed to normalize args: %s", err))
			}
		}
		plan.Queries = append(plan.Queries, &analyzed)
	}
	return plan, planErr.wrap()
}

func (a *analyzer) analyzeQuery(node sql_parser.SQLNode) (domains.CachePlanQuery, error) {
	q := newQueryAnalyzer(a.schemas)
	switch n := node.(type) {
	case sql_parser.SelectStmtNode:
		return q.analyzeSelectStmt(n)
	case sql_parser.InsertStmtNode:
		return q.analyzeInsertStmt(n)
	case sql_parser.UpdateStmtNode:
		return q.analyzeUpdateStmt(n)
	case sql_parser.DeleteStmtNode:
		return q.analyzeDeleteStmt(n)
	default:
		return domains.CachePlanQuery{}, fmt.Errorf("unknown query type: %T", node)
	}
}

func (a *analyzer) normalizeArgs(sql string, queryPlan *domains.CachePlanQuery) error {
	result, err := normalizer.NormalizeArgs(sql)
	if err != nil {
		return fmt.Errorf("failed to normalize args: %s", err)
	}
	queryPlan.Query = result.Query
	switch queryPlan.Type {
	case domains.CachePlanQueryType_SELECT:
		for i, arg := range result.ExtraArgs {
			queryPlan.Select.Conditions = append(queryPlan.Select.Conditions, domains.CachePlanCondition{
				Column:      arg.Column,
				Operator:    domains.CachePlanOperator_EQ,
				Placeholder: domains.CachePlanPlaceholder{Index: i, Extra: true},
			})
		}
	case domains.CachePlanQueryType_UPDATE:
		for i, set := range result.ExtraSets {
			queryPlan.Update.Targets = append(queryPlan.Update.Targets, domains.CachePlanUpdateTarget{
				Column:      set.Column,
				Placeholder: domains.CachePlanPlaceholder{Index: i, Extra: true},
			})
		}
		for i, arg := range result.ExtraArgs {
			queryPlan.Update.Conditions = append(queryPlan.Update.Conditions, domains.CachePlanCondition{
				Column:      arg.Column,
				Operator:    domains.CachePlanOperator_EQ,
				Placeholder: domains.CachePlanPlaceholder{Index: i + len(result.ExtraSets), Extra: true},
			})
		}
	case domains.CachePlanQueryType_DELETE:
		for i, arg := range result.ExtraArgs {
			queryPlan.Delete.Conditions = append(queryPlan.Delete.Conditions, domains.CachePlanCondition{
				Column:      arg.Column,
				Operator:    domains.CachePlanOperator_EQ,
				Placeholder: domains.CachePlanPlaceholder{Index: i, Extra: true},
			})
		}
	}

	return nil
}
