package dynamic_extractor

import "sync"

var queries sync.Map

func deleteAllQueries() {
	queries.Clear()
}

func addQuery(query string) {
	queries.Store(query, struct{}{})
}

func getQueries() []string {
	ret := make([]string, 0)
	for query := range queries.Range {
		ret = append(ret, query.(string))
	}
	return ret
}
