package limiter_test

import (
	"testing"
	"time"

	"github.com/JhayceeCodes/seigen/internal/limiter"
)

func TestTokenBucketStartsFull(t *testing.T) {

	bucket := limiter.NewTokenBucket(
		5,
		time.Second,
		1,
	)

	if bucket.Tokens() != 5 {
		t.Errorf("expected 5 tokens, got %d", bucket.Tokens())
	}

}

func TestTokenBucketAllowConsumesOneToken(t *testing.T) {

	bucket := limiter.NewTokenBucket(
		5,
		time.Second,
		1,
	)

	bucket.Allow()

	if bucket.Tokens() != 4 {
		t.Errorf(
			"expected 4 tokens, got %d",
			bucket.Tokens(),
		)
	}

}

func TestTokenBucketRejectWhenEmpty(t *testing.T) {

	bucket := limiter.NewTokenBucket(
		1,
		time.Second,
		1,
	)

	bucket.Allow()

	allowed := bucket.Allow()

	if allowed {
		t.Error("expected request to be rejected")
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

	if !bucket.Allow() {
		t.Error("expected token to be refilled")
	}

}




func TestTokenBucketAllowNConsumesRequestedTokens(t *testing.T) {
	bucket := limiter.NewTokenBucket(10, time.Second, 1)

	if !bucket.AllowN(4) {
		t.Fatal("expected request to be allowed")
	}

	if bucket.Tokens() != 6 {
		t.Errorf("expected 6 tokens, got %d", bucket.Tokens())
	}
}


func TestTokenBucketAllowNRejectsWhenInsufficientTokens(t *testing.T) {
	bucket := limiter.NewTokenBucket(5, time.Second, 1)

	if bucket.AllowN(6) {
		t.Fatal("expected request to be rejected")
	}

	if bucket.Tokens() != 5 {
		t.Errorf("expected tokens to remain unchanged, got %d", bucket.Tokens())
	}
}


func TestAllowNRejectsInvalidAmount(t *testing.T) {
	bucket := limiter.NewTokenBucket(5, time.Second, 1)

	if bucket.AllowN(0) {
		t.Error("expected zero-cost request to be rejected")
	}

	if bucket.AllowN(-1) {
		t.Error("expected negative-cost request to be rejected")
	}
}