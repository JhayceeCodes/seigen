package store_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/JhayceeCodes/seigen/model"
	"github.com/JhayceeCodes/seigen/store"
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



func TestNewInMemoryPolicyGroupRepositoryStartsEmpty(t *testing.T) {
	groupStore := store.NewInMemoryPolicyGroupRepository()

	_, err := groupStore.Get("free")

	if !errors.Is(err, store.ErrPolicyGroupNotFound) {
		t.Fatalf("expected ErrPolicyGroupNotFound, got %v", err)
	}
}

func TestSetStoresAPolicyGroup(t *testing.T) {
	group := model.PolicyGroup{
		Name: "free",
		Limiter: model.LimiterConfig{
			Algorithm: model.FixedWindow,
			Config: model.WindowConfig{
				Limit:  10,
				Window: time.Minute,
			},
		},
	}

	groupStore := store.NewInMemoryPolicyGroupRepository()

	if err := groupStore.Set(group); err != nil {
		t.Fatalf("expected policy group to be stored, got error: %v", err)
	}

	got, err := groupStore.Get(group.Name)
	if err != nil {
		t.Fatalf("expected policy group to be retrieved, got error: %v", err)
	}

	if got.Name != group.Name {
		t.Errorf("expected group name %q, got %q",
			group.Name, got.Name)
	}

	if got.Limiter.Algorithm != group.Limiter.Algorithm {
		t.Errorf("expected algorithm %q, got %q",
			group.Limiter.Algorithm, got.Limiter.Algorithm)
	}

	config := got.Limiter.Config.(model.WindowConfig)

	if config.Limit != 10 {
		t.Errorf("expected limit 10, got %d", config.Limit)
	}
}

func TestGetPolicyGroupReturnsNotFound(t *testing.T) {
	groupStore := store.NewInMemoryPolicyGroupRepository()

	_, err := groupStore.Get("unknown")

	if !errors.Is(err, store.ErrPolicyGroupNotFound) {
		t.Fatalf("expected ErrPolicyGroupNotFound, got %v", err)
	}
}

func TestDeleteRemovesPolicyGroup(t *testing.T) {
	group := model.PolicyGroup{
		Name: "free",
		Limiter: model.LimiterConfig{
			Algorithm: model.FixedWindow,
			Config: model.WindowConfig{
				Limit:  10,
				Window: time.Minute,
			},
		},
	}

	groupStore := store.NewInMemoryPolicyGroupRepository()

	if err := groupStore.Set(group); err != nil {
		t.Fatalf("expected policy group to be stored, got error: %v", err)
	}

	if err := groupStore.Delete(group.Name); err != nil {
		t.Fatalf("expected policy group to be deleted, got error: %v", err)
	}

	_, err := groupStore.Get(group.Name)

	if !errors.Is(err, store.ErrPolicyGroupNotFound) {
		t.Fatalf("expected ErrPolicyGroupNotFound, got %v", err)
	}
}

func TestDeletePolicyGroupReturnsNotFound(t *testing.T) {
	groupStore := store.NewInMemoryPolicyGroupRepository()

	err := groupStore.Delete("unknown")

	if !errors.Is(err, store.ErrPolicyGroupNotFound) {
		t.Fatalf("expected ErrPolicyGroupNotFound, got %v", err)
	}
}

func TestSetRejectsInvalidPolicyGroup(t *testing.T) {
	group := model.PolicyGroup{
		Name: "free",
		Limiter: model.LimiterConfig{
			Algorithm: model.FixedWindow,
			Config: model.TokenBucketConfig{
				Capacity:       10,
				RefillInterval: time.Second,
				RefillAmount:   1,
			},
		},
	}

	groupStore := store.NewInMemoryPolicyGroupRepository()

	if err := groupStore.Set(group); err == nil {
		t.Fatal("expected invalid policy group to be rejected")
	}
}

func TestSetReplacesExistingPolicyGroup(t *testing.T) {
	groupStore := store.NewInMemoryPolicyGroupRepository()

	group := model.PolicyGroup{
		Name: "free",
		Limiter: model.LimiterConfig{
			Algorithm: model.FixedWindow,
			Config: model.WindowConfig{
				Limit:  10,
				Window: time.Minute,
			},
		},
	}

	if err := groupStore.Set(group); err != nil {
		t.Fatalf("expected policy group to be stored, got error: %v", err)
	}

	updatedGroup := model.PolicyGroup{
		Name: "free",
		Limiter: model.LimiterConfig{
			Algorithm: model.FixedWindow,
			Config: model.WindowConfig{
				Limit:  100,
				Window: time.Minute,
			},
		},
	}

	if err := groupStore.Set(updatedGroup); err != nil {
		t.Fatalf("expected updated policy group to be stored, got error: %v", err)
	}

	got, err := groupStore.Get("free")
	if err != nil {
		t.Fatalf("expected policy group to exist, got error: %v", err)
	}

	config := got.Limiter.Config.(model.WindowConfig)

	if config.Limit != 100 {
		t.Errorf("expected updated limit 100, got %d", config.Limit)
	}
}