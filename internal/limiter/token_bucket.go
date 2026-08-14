package limiter

import (
	"sync"
	"time"
)

type TokenBucket struct {
	capacity       int
	tokens         int
	refillAmount   int
	refillInterval time.Duration
	lastRefill     time.Time
	mu             sync.Mutex
}

// NewTokenBucket creates a new token bucket rate limiter.
//
// capacity defines the maximum number of tokens the bucket can hold.
//
// interval specifies how often tokens are replenished.
//
// amount specifies how many tokens are added every refill interval.
// If amount is less than or equal to zero, it defaults to 1.
//
// Refills are computed lazily based on elapsed time whenever the bucket
// is accessed, rather than by a background goroutine.
func NewTokenBucket(capacity int, interval time.Duration, amount int) *TokenBucket {
	if capacity <= 0 {
		panic("capacity cannot be less than zero")
	}

	if interval <= 0 {
		panic("interval cannot be less than zero")
	}

	if amount <= 0 {
		amount = 1
	}

	return &TokenBucket{
		capacity:       capacity,
		tokens:         capacity,
		refillAmount:   amount,
		refillInterval: interval,
		lastRefill:     time.Now(),
	}
}

func (tb *TokenBucket) refill() {
	elapsed := time.Since(tb.lastRefill)
	intervalsPassed := int(elapsed / tb.refillInterval)

	if intervalsPassed <= 0 {
		return
	}

	tb.tokens += intervalsPassed * tb.refillAmount
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	tb.lastRefill = tb.lastRefill.Add(time.Duration(intervalsPassed) * tb.refillInterval)
}

func (tb *TokenBucket) Allow() bool {
	return tb.AllowN(1)
}

func (tb *TokenBucket) AllowN(amount int) bool {
	if amount <= 0 {
		return false
	}

	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens < amount {
		return false
	}

	tb.tokens -= amount
	return true
}

func (tb *TokenBucket) Tokens() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()
	return tb.tokens
}

func (tb *TokenBucket) LastRefill() time.Time {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	return tb.lastRefill
}
