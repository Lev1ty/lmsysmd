package static

import (
	"log"
	"net/http"
	"strconv"
	"time"
)

var buildTimestampSeconds string

func Middleware(next http.Handler) http.Handler {
	s, err := strconv.ParseInt(buildTimestampSeconds, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	t := time.Unix(s, 0)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Last-Modified", t.Format(http.TimeFormat))
		w.Header().Set("Document-Policy", "js-profiling")
		next.ServeHTTP(w, r)
	})
}
