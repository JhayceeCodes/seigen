package limiter_test

import (
	"testing"
	"time"

	"github.com/JhayceeCodes/seigen/limiter"
)

func TestTokenBucketStartsFull(t *testing.T) {
	bucket := limiter.NewTokenBucket(
		5,
		time.Second,
		1,
	)

	if got := bucket.Remaining(); got != 5 {
		t.Errorf("expected 5 remaining tokens, got %d", got)
	}
}

func TestTokenBucketAllowConsumesOneToken(t *testing.T) {
	bucket := limiter.NewTokenBucket(
		5,
		time.Second,
		1,
	)

	result := bucket.Allow()

	if !result.Allowed {
		t.Fatal("expected request to be allowed")
	}

	if result.Remaining != 4 {
		t.Errorf("expected 4 remaining tokens, got %d", result.Remaining)
	}

	if result.RetryAfter != 0 {
		t.Errorf("expected no retry delay, got %v", result.RetryAfter)
	}
}

func TestTokenBucketRejectWhenEmpty(t *testing.T) {
	bucket := limiter.NewTokenBucket(
		1,
		time.Second,
		1,
	)

	bucket.Allow()

	result := bucket.Allow()

	if result.Allowed {
		t.Error("expected request to be rejected")
	}

	if result.Remaining != 0 {
		t.Errorf("expected 0 remaining tokens, got %d", result.Remaining)
	}

	if result.RetryAfter <= 0 {
		t.Error("expected a positive retry delay")
	}
}

func TestTokenBucketRefill(t *testing.T) {
	bucket := limiter.NewTokenBucket(
		2,
		time.Second,
		1,
	)

	bucket.Allow()
	bucket.Allow()

	time.Sleep(1100 * time.Millisecond)

	result := bucket.Allow()

	if !result.Allowed {
		t.Error("expected token to be refilled")
	}

	if result.Remaining != 0 {
		t.Errorf("expected 0 remaining tokens after consuming refilled token, got %d", result.Remaining)
	}
}

func TestTokenBucketAllowNConsumesRequestedTokens(t *testing.T) {
	bucket := limiter.NewTokenBucket(
		10,
		time.Second,
		1,
	)

	result := bucket.AllowN(4)

	if !result.Allowed {
		t.Fatal("expected request to be allowed")
	}

	if result.Remaining != 6 {
		t.Errorf("expected 6 remaining tokens, got %d", result.Remaining)
	}

	if result.RetryAfter != 0 {
		t.Errorf("expected no retry delay, got %v", result.RetryAfter)
	}
}

func TestTokenBucketAllowNRejectsWhenInsufficientTokens(t *testing.T) {
	bucket := limiter.NewTokenBucket(
		5,
		time.Second,
		1,
	)

	result := bucket.AllowN(6)

	if result.Allowed {
		t.Fatal("expected request to be rejected")
	}

	if result.Remaining != 5 {
		t.Errorf("expected 5 remaining tokens, got %d", result.Remaining)
	}

	if result.RetryAfter != time.Second {
		t.Errorf(
			"expected retry after %v, got %v",
			time.Second,
			result.RetryAfter,
		)
	}
}

func TestAllowNRejectsInvalidAmount(t *testing.T) {
	bucket := limiter.NewTokenBucket(
		5,
		time.Second,
		1,
	)

	result := bucket.AllowN(0)

	if result.Allowed {
		t.Error("expected zero-cost request to be rejected")
	}

	result = bucket.AllowN(-1)

	if result.Allowed {
		t.Error("expected negative-cost request to be rejected")
	}
}

func TestTokenBucketAllowNCalculatesRetryAfter(t *testing.T) {
	bucket := limiter.NewTokenBucket(
		10,
		500*time.Millisecond,
		2,
	)

	// 10 tokens initially. Consume 8, leaving 2.
	bucket.AllowN(8)

	// Request needs 5 tokens, so 3 more are needed.
	// Each interval provides 2 tokens:
	//
	// 1st interval → 4 tokens
	// 2nd interval → 6 tokens
	//
	// Therefore, the request can succeed after 2 intervals.
	result := bucket.AllowN(5)

	if result.Allowed {
		t.Fatal("expected request to be rejected")
	}

	expected := time.Second

	if result.RetryAfter != expected {
		t.Errorf(
			"expected retry after %v, got %v",
			expected,
			result.RetryAfter,
		)
	}

	if result.Remaining != 2 {
		t.Errorf(
			"expected 2 remaining tokens, got %d",
			result.Remaining,
		)
	}
}
