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
	mu             sync.RWMutex
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
// A background goroutine is started to replenish tokens periodically.
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

	tb := &TokenBucket{
		capacity:       capacity,
		refillAmount:   amount,
		refillInterval: interval,
	}

	tb.tokens = capacity
	tb.lastRefill = time.Now()

	go tb.startRefill()

	return tb
}

func (tb *TokenBucket) startRefill() {
	ticker := time.NewTicker(tb.refillInterval)
	defer ticker.Stop()

	for t := range ticker.C {
		tb.mu.Lock()

		if tb.tokens < tb.capacity {
			tb.tokens += tb.refillAmount

			if tb.tokens > tb.capacity {
				tb.tokens = tb.capacity
			}

			tb.lastRefill = t
		}

		tb.mu.Unlock()
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if tb.tokens == 0 {
		return false
	}

	tb.tokens--
	return true

}

func (tb *TokenBucket) Tokens() int {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	return tb.tokens
}

func (tb *TokenBucket) LastRefill() time.Time {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	return tb.lastRefill
}
