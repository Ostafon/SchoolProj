package middlewares

import (
	"net/http"
	"strings"
)

func MiddlewaresExcludeRoute(middleware func(http.Handler) http.Handler, excPath ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, path := range excPath {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			middleware(next).ServeHTTP(w, r)
		})
	}

}
