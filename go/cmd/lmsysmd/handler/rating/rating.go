package rating

import (
	"fmt"
	"net/http"
	"time"
)

func PatternAndHandler(middleware func(http.Handler) http.Handler) (string, http.Handler) {
	return "GET /rating/{$}", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, fmt.Sprintf("/rating/id?id=1&ts=%d", time.Now().Unix()), http.StatusTemporaryRedirect)
	}))
}
