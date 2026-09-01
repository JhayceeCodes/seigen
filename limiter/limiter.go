package limiter

import "time"

type LimiterResult struct {
	Allowed    bool
	Remaining  int
	RetryAfter time.Duration
}

type Limiter interface {
	Allow() LimiterResult
}
