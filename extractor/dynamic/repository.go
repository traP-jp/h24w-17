package dynamic_extractor

var queries = map[string]struct{}{}

func deleteAllQueries() {
	queries = make(map[string]struct{})
}

func addQuery(query string) {
	queries[query] = struct{}{}
}

func getQueries() []string {
	ret := make([]string, 0, len(queries))
	for query := range queries {
		ret = append(ret, query)
	}
	return ret
}
