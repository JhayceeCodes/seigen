package store

import (
	"sync"

	"github.com/JhayceeCodes/seigen/model"
)

type InMemoryPolicyRepository struct {
	mu       sync.RWMutex
	policies map[model.Identifier]model.Policy
}

func NewInMemoryPolicyRepository() *InMemoryPolicyRepository {
	return &InMemoryPolicyRepository{
		policies: make(map[model.Identifier]model.Policy),
	}
}

func (r *InMemoryPolicyRepository) Set(policy model.Policy) error {
	if err := policy.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.policies[policy.Identifier] = policy

	return nil
}

func (r *InMemoryPolicyRepository) Get(identifier model.Identifier) (model.Policy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	policy, ok := r.policies[identifier]
	if !ok {
		return model.Policy{}, ErrPolicyNotFound
	}

	return policy, nil
}

func (r *InMemoryPolicyRepository) Delete(identifier model.Identifier) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.policies[identifier]
	if !ok {
		return ErrPolicyNotFound
	}

	delete(r.policies, identifier)

	return nil
}
