package middlewares

import (
	"net/http"
	"sync"
	"time"
)

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]int
	limit    int
	reset    time.Duration
}

func NewRateLimiter(limit int, reset time.Duration) *rateLimiter {
	rl := &rateLimiter{visitors: make(map[string]int), limit: limit, reset: reset}
	go rl.resetVisitors()
	return rl
}

func (rl *rateLimiter) resetVisitors() {
	for {
		time.Sleep(rl.reset)
		rl.mu.Lock()
		rl.visitors = make(map[string]int)
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rl.mu.Lock()
		defer rl.mu.Unlock()
		visitorIP := r.RemoteAddr
		rl.visitors[visitorIP]++

		if rl.visitors[visitorIP] > rl.limit {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
