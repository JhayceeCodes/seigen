package limiter_test

import (
	"sync"
	"testing"
	"time"

	"github.com/JhayceeCodes/seigen/internal/limiter"
	"github.com/JhayceeCodes/seigen/internal/model"
)

func TestManagerReturnsSameLimiterForIdentifier(t *testing.T) {
	manager := limiter.NewManager()

	config := model.LimiterConfig{
		Algorithm: model.TokenBucket,
		Config: model.TokenBucketConfig{
			Capacity:       2,
			RefillInterval: time.Second,
			RefillAmount:   1,
		},
	}

	first, err := manager.Get("user:123", config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	second, err := manager.Get("user:123", config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first != second {
		t.Fatal("expected same limiter instance")
	}
}

func TestManagerCreatesSeparateLimiters(t *testing.T) {
	manager := limiter.NewManager()

	config := model.LimiterConfig{
		Algorithm: model.TokenBucket,
		Config: model.TokenBucketConfig{
			Capacity:       2,
			RefillInterval: time.Second,
			RefillAmount:   1,
		},
	}

	first, err := manager.Get("user:123", config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	second, err := manager.Get("user:456", config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first == second {
		t.Fatal("expected different limiter instances for different identifiers")
	}
}

func TestManagerConcurrentAccess(t *testing.T) {
	manager := limiter.NewManager()

	config := model.LimiterConfig{
		Algorithm: model.TokenBucket,
		Config: model.TokenBucketConfig{
			Capacity:       10,
			RefillInterval: time.Second,
			RefillAmount:   1,
		},
	}

	const goroutines = 20

	results := make(chan limiter.Limiter, goroutines)

	var wg sync.WaitGroup

	for range goroutines {
		wg.Go(func() {
			got, err := manager.Get("user:123", config)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			results <- got
		})
	}

	wg.Wait()
	close(results)

	var first limiter.Limiter
	count := 0

	for got := range results {
		count++

		if first == nil {
			first = got
			continue
		}

		if got != first {
			t.Fatal("expected all goroutines to receive the same limiter instance")
		}
	}

	if count != goroutines {
		t.Fatalf("expected %d results, got %d", goroutines, count)
	}
}

func TestManagerRejectsInvalidConfig(t *testing.T) {
	manager := limiter.NewManager()

	config := model.LimiterConfig{
		Algorithm: model.TokenBucket,
		Config: model.WindowConfig{
			Limit:  3,
			Window: 5 * time.Second,
		},
	}

	if _, err := manager.Get("user:123", config); err == nil {
		t.Fatal("expected invalid limiter configuration to be rejected")
	}

}

func TestManagerReplacesLimiterWhenPolicyChanges(t *testing.T) {
	manager := limiter.NewManager()

	firstConfig := model.LimiterConfig{
		Algorithm: model.TokenBucket,
		Config: model.TokenBucketConfig{
			Capacity:       2,
			RefillInterval: time.Second,
			RefillAmount:   1,
		},
	}

	secondConfig := model.LimiterConfig{
		Algorithm: model.TokenBucket,
		Config: model.TokenBucketConfig{
			Capacity:       10,
			RefillInterval: time.Second,
			RefillAmount:   1,
		},
	}

	first, err := manager.Get("user:123", firstConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	second, err := manager.Get("user:123", secondConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first == second {
		t.Fatal("expected new limiter when policy changes")
	}
}
