package middlewares

import (
	"log"
	"net/http"
	"strings"
)

type HPPOptions struct {
	CheckQuery          bool
	CheckBody           bool
	CheckOnlyForContent string
	Whitelist           []string
}

func Hpp(hpp HPPOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if hpp.CheckBody && r.Method == http.MethodPost && isCorrectContentType(r, hpp.CheckOnlyForContent) {
				filterBodyParams(r, hpp.Whitelist)
			}
			if hpp.CheckQuery && r.URL.Query() != nil {
				filterQueryParams(r, hpp.Whitelist)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isCorrectContentType(r *http.Request, contentType string) bool {
	return strings.Contains(r.Header.Get("Content-Type"), contentType)
}

func filterBodyParams(r *http.Request, whitelist []string) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}
	for k, v := range r.Form {
		if len(v) > 1 {
			r.Form.Set(k, v[0])
		}
		if isWhitelist(k, whitelist) {
			delete(r.Form, k)
		}
	}
}

func isWhitelist(param string, whitelist []string) bool {
	for _, v := range whitelist {
		if v == param {
			return true
		}
	}
	return false
}

func filterQueryParams(r *http.Request, whitelist []string) {
	q := r.URL.Query()
	for k, v := range q {
		if len(v) > 1 {
			q.Set(k, v[0])
		}
		if isWhitelist(k, whitelist) {
			q.Del(k)
		}
	}
	r.URL.RawQuery = q.Encode()
}
