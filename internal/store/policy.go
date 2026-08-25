package store

import (
	"sync"

	"github.com/JhayceeCodes/seigen/internal/model"
)

type PolicyStore struct {
	mu       sync.RWMutex
	policies map[model.Identifier]model.Policy
}

func NewPolicyStore() *PolicyStore {
	return &PolicyStore{
		policies: make(map[model.Identifier]model.Policy),
	}
}

func (ps *PolicyStore) Set(policy model.Policy) error {
	if err := policy.Validate(); err != nil {
		return err
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.policies[policy.Identifier] = policy

	return nil
}

func (ps *PolicyStore) Get(identifier model.Identifier) (model.Policy, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	policy, ok := ps.policies[identifier]
	if !ok {
		return model.Policy{}, ErrPolicyNotFound
	}

	return policy, nil
}

func (ps *PolicyStore) Delete(identifier model.Identifier) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	_, ok := ps.policies[identifier]
	if !ok {
		return ErrPolicyNotFound
	}

	delete(ps.policies, identifier)

	return nil
}
