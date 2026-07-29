package limiter

import (
	"testing"
	"time"
)

func TestBucketStartsFull(t *testing.T) {

	bucket := NewTokenBucket(
		5,
		time.Second,
		1,
	)

	if bucket.tokens != 5 {
		t.Errorf("expected 5 tokens, got %d", bucket.tokens)
	}

}

func TestAllowConsumesOneToken(t *testing.T) {

	bucket := NewTokenBucket(
		5,
		time.Second,
		1,
	)

	bucket.Allow()

	if bucket.tokens != 4 {
		t.Errorf(
			"expected 4 tokens, got %d",
			bucket.tokens,
		)
	}

}

func TestRejectWhenEmpty(t *testing.T) {

	bucket := NewTokenBucket(
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

func TestRefill(t *testing.T) {

	bucket := NewTokenBucket(
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
