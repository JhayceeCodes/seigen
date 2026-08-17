package middleware

import (
	"net/http"

	"github.com/JhayceeCodes/seigen/internal/limiter"
)

func RateLimit(limiter limiter.Limiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
