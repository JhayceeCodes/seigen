package limiter

import "time"

// LimiterResult contains the outcome of a rate-limit evaluation.
//
// Allowed indicates whether the request may proceed. Remaining reports the
// estimated number of requests that can still be accepted, while RetryAfter
// indicates how long the caller should wait when the request is rejected.
type LimiterResult struct {
	Allowed    bool
	Remaining  int
	RetryAfter time.Duration
}

// Limiter defines the interface implemented by all rate-limiting algorithms.
//
// Allow evaluates whether the next request should be permitted.
type Limiter interface {
	Allow() LimiterResult
}
