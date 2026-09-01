package middleware

import (
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/JhayceeCodes/seigen/service"
	"github.com/JhayceeCodes/seigen/store"
)

func RateLimit(
	rateLimitService *service.RateLimitService,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result, err := rateLimitService.Evaluate(r)
		if err != nil {
			if errors.Is(err, store.ErrPolicyNotFound) {
				http.Error(w, "rate limit policy not configured", http.StatusInternalServerError)
				return
			}

			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set(
			"X-RateLimit-Limit",
			strconv.Itoa(result.Limit),
		)

		w.Header().Set(
			"X-RateLimit-Remaining",
			strconv.Itoa(result.Remaining),
		)

		if !result.Allowed {
			retryAfter := int(math.Ceil(result.RetryAfter.Seconds()))

			w.Header().Set(
				"Retry-After",
				strconv.Itoa(retryAfter),
			)

			http.Error(
				w,
				"too many requests",
				http.StatusTooManyRequests,
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}
