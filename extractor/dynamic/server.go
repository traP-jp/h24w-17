package dynamic_extractor

import (
	"log"
	"net/http"
	"os"
)

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		res := ""
		queries := getQueries()
		for _, query := range queries {
			res += query + "\n"
		}
		w.Write([]byte(res))
		w.WriteHeader(http.StatusOK)
	})

	port := os.Getenv("EXTRACTOR_PORT")
	if port == "" {
		port = "39393"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
