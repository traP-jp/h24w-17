package h24w17

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/traP-jp/h24w-17/domains"
)

type NormalizedQuery struct {
	Query           string
	ExtraConditions []domains.CachePlanCondition
}

func unwrapInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Errorf("failed to unwrap int: %w", err))
	}
	return i
}

var patterns = []replacementPattern{
	{
		regex:       regexp.MustCompile(`(?i)\b([a-z]+)\s*=\s*(\d+)`),
		replacement: "$1 = ?",
		condition: func(match []string) *domains.CachePlanCondition {
			return &domains.CachePlanCondition{Column: match[1], Value: unwrapInt(match[2])}
		},
	},
	{
		regex:       regexp.MustCompile(`(?i)\b([a-z]+)\s*=\s*['"]([^']*)['"]`),
		replacement: "$1 = ?",
		condition: func(match []string) *domains.CachePlanCondition {
			return &domains.CachePlanCondition{Column: match[1], Value: match[2]}
		},
	},
	{
		regex:       regexp.MustCompile(`(?i)\b(IN|VALUES)\s*\(\s*\?\s*(?:\s*,\s*\?)*\s*\)`),
		replacement: "$1 (?)",
		condition: func(match []string) *domains.CachePlanCondition {
			return nil
		},
	},
	{
		regex:       regexp.MustCompile(`(?i)\bLIMIT\s+(\d+)`),
		replacement: "LIMIT ?",
		condition: func(match []string) *domains.CachePlanCondition {
			return &domains.CachePlanCondition{Column: "LIMIT()", Value: unwrapInt(match[1])}
		},
	},
}

type replacementPattern struct {
	regex       *regexp.Regexp
	replacement string
	condition   func(match []string) *domains.CachePlanCondition
}

func NormalizeQuery(query string) (NormalizedQuery, error) {
	if query == "" {
		return NormalizedQuery{}, errors.New("query cannot be empty")
	}

	extraConditions := []domains.CachePlanCondition{}

	normalizedQuery := query

	// Apply patterns to normalize the query and collect conditions
	for _, pattern := range patterns {
		matches := pattern.regex.FindAllStringSubmatch(normalizedQuery, -1)
		for _, match := range matches {
			normalizedQuery = pattern.regex.ReplaceAllString(normalizedQuery, pattern.replacement)
			if pattern.condition != nil {
				condition := pattern.condition(match)
				if condition != nil {
					extraConditions = append(extraConditions, *condition)
				}
			}
		}
	}

	// Ensure consistent spacing and formatting
	normalizedQuery = strings.TrimSpace(normalizedQuery)

	return NormalizedQuery{
		Query:           normalizedQuery,
		ExtraConditions: extraConditions,
	}, nil
}
