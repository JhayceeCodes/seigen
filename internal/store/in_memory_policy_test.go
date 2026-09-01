package store_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/JhayceeCodes/seigen/internal/model"
	"github.com/JhayceeCodes/seigen/internal/store"
)

func TestNewInMemoryPolicyRepositoryStartsEmpty(t *testing.T) {
	policyStore := store.NewInMemoryPolicyRepository()

	_, err := policyStore.Get("user:123")

	if !errors.Is(err, store.ErrPolicyNotFound) {
		t.Fatalf("expected ErrPolicyNotFound, got %v", err)
	}
}

func TestSetStoresAPolicy(t *testing.T) {
	policy := model.Policy{
		Identifier: "user:123",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.TokenBucketConfig{
				Capacity:       10,
				RefillInterval: time.Second,
				RefillAmount:   1,
			},
		},
	}

	policyStore := store.NewInMemoryPolicyRepository()

	if err := policyStore.Set(policy); err != nil {
		t.Fatalf("expected policy to be stored, got error: %v", err)
	}

	got, err := policyStore.Get(policy.Identifier)
	if err != nil {
		t.Fatalf("expected policy to be retrieved, got error: %v", err)
	}

	if got.Identifier != policy.Identifier {
		t.Errorf("expected identifier %q, got %q",
			policy.Identifier, got.Identifier)
	}

	if got.Limiter.Algorithm != policy.Limiter.Algorithm {
		t.Errorf("expected algorithm %q, got %q",
			policy.Limiter.Algorithm, got.Limiter.Algorithm)
	}
}

func TestGetReturnsNotFoundForUnknownIdentifier(t *testing.T) {
	policyStore := store.NewInMemoryPolicyRepository()

	_, err := policyStore.Get("user:unknown")

	if !errors.Is(err, store.ErrPolicyNotFound) {
		t.Fatalf("expected ErrPolicyNotFound, got %v", err)
	}
}

func TestDeleteRemovesPolicy(t *testing.T) {
	policy := model.Policy{
		Identifier: "user:123",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.TokenBucketConfig{
				Capacity:       10,
				RefillInterval: time.Second,
				RefillAmount:   1,
			},
		},
	}

	policyStore := store.NewInMemoryPolicyRepository()

	if err := policyStore.Set(policy); err != nil {
		t.Fatalf("expected policy to be stored, got error: %v", err)
	}

	if err := policyStore.Delete(policy.Identifier); err != nil {
		t.Fatalf("expected policy to be deleted, got error: %v", err)
	}

	_, err := policyStore.Get(policy.Identifier)

	if !errors.Is(err, store.ErrPolicyNotFound) {
		t.Fatalf("expected ErrPolicyNotFound, got %v", err)
	}
}

func TestDeleteReturnsNotFound(t *testing.T) {
	policyStore := store.NewInMemoryPolicyRepository()

	err := policyStore.Delete("user:unknown")

	if !errors.Is(err, store.ErrPolicyNotFound) {
		t.Fatalf("expected ErrPolicyNotFound, got %v", err)
	}
}

func TestSetRejectsInvalidPolicy(t *testing.T) {
	policy := model.Policy{
		Identifier: "user:123",
		Limiter: model.LimiterConfig{
			Algorithm: model.FixedWindow,
			Config: model.TokenBucketConfig{
				Capacity:       10,
				RefillInterval: time.Second,
				RefillAmount:   1,
			},
		},
	}

	policyStore := store.NewInMemoryPolicyRepository()

	if err := policyStore.Set(policy); err == nil {
		t.Fatal("expected invalid policy to be rejected")
	}
}

func TestSetReplacesExistingPolicy(t *testing.T) {
	policyStore := store.NewInMemoryPolicyRepository()

	policy := model.Policy{
		Identifier: "user:123",
		Limiter: model.LimiterConfig{
			Algorithm: model.FixedWindow,
			Config: model.WindowConfig{
				Limit:  10,
				Window: time.Minute,
			},
		},
	}

	if err := policyStore.Set(policy); err != nil {
		t.Fatalf("expected policy to be stored, got error: %v", err)
	}

	updatedPolicy := model.Policy{
		Identifier: "user:123",
		Limiter: model.LimiterConfig{
			Algorithm: model.FixedWindow,
			Config: model.WindowConfig{
				Limit:  100,
				Window: time.Minute,
			},
		},
	}

	if err := policyStore.Set(updatedPolicy); err != nil {
		t.Fatalf("expected updated policy to be stored, got error: %v", err)
	}

	got, err := policyStore.Get("user:123")
	if err != nil {
		t.Fatalf("expected policy to exist, got error: %v", err)
	}

	config := got.Limiter.Config.(model.WindowConfig)

	if config.Limit != 100 {
		t.Errorf("expected updated limit 100, got %d", config.Limit)
	}
}

func TestPolicyStoreConcurrentAccess(t *testing.T) {
	policyStore := store.NewInMemoryPolicyRepository()

	policy := model.Policy{
		Identifier: "user:123",
		Limiter: model.LimiterConfig{
			Algorithm: model.FixedWindow,
			Config: model.WindowConfig{
				Limit:  100,
				Window: time.Minute,
			},
		},
	}

	var wg sync.WaitGroup

	for range 100 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if err := policyStore.Set(policy); err != nil {
				t.Errorf("unexpected Set error: %v", err)
			}

			_, err := policyStore.Get(policy.Identifier)
			if err != nil {
				t.Errorf("unexpected Get error: %v", err)
			}
		}()
	}

	wg.Wait()
}

func TestPolicyStoreConcurrentWrites(t *testing.T) {
	policyStore := store.NewInMemoryPolicyRepository()

	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			policy := model.Policy{
				Identifier: model.Identifier(fmt.Sprintf("user:%d", i)),
				Limiter: model.LimiterConfig{
					Algorithm: model.FixedWindow,
					Config: model.WindowConfig{
						Limit:  100,
						Window: time.Minute,
					},
				},
			}

			if err := policyStore.Set(policy); err != nil {
				t.Errorf("unexpected Set error: %v", err)
			}
		}(i)
	}

	wg.Wait()
}
