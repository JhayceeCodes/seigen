package store

import (
	"sync"

	"github.com/JhayceeCodes/seigen/model"
)

type InMemoryPolicyRepository struct {
	mu       sync.RWMutex
	policies map[model.Identifier]model.Policy
}

type InMemoryPolicyGroupRepository struct {
	mu     sync.RWMutex
	groups map[string]model.PolicyGroup
	ids    map[int64]string
}

type InMemoryPolicyGroupMemberRepository struct {
	mu      sync.RWMutex
	members map[model.Identifier]string
	groups  PolicyGroupRepository
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

// Policy groups

func NewInMemoryPolicyGroupRepository() *InMemoryPolicyGroupRepository {
    return &InMemoryPolicyGroupRepository{
        groups: make(map[string]model.PolicyGroup),
        ids:    make(map[int64]string),
    }
}

func (g *InMemoryPolicyGroupRepository) Set(group model.PolicyGroup) error {
	if err := group.Validate(); err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	g.groups[group.Name] = group

	return nil
}

func (g *InMemoryPolicyGroupRepository) Get(name string) (model.PolicyGroup, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	group, ok := g.groups[name]
	if !ok {
		return model.PolicyGroup{}, ErrPolicyGroupNotFound
	}

	return group, nil
}

func (g *InMemoryPolicyGroupRepository) Delete(name string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	_, ok := g.groups[name]
	if !ok {
		return ErrPolicyGroupNotFound
	}

	delete(g.groups, name)

	return nil
}

// Policy group membership

func NewInMemoryPolicyGroupMemberRepository(
	groups PolicyGroupRepository,
) *InMemoryPolicyGroupMemberRepository {
	return &InMemoryPolicyGroupMemberRepository{
		members: make(map[model.Identifier]string),
		groups:  groups,
	}
}

func (r *InMemoryPolicyGroupMemberRepository) GetGroup(
	identifier model.Identifier,
) (model.PolicyGroup, error) {
	r.mu.RLock()
	groupName, ok := r.members[identifier]
	r.mu.RUnlock()

	if !ok {
		return model.PolicyGroup{}, ErrPolicyGroupMemberNotFound
	}

	return r.groups.Get(groupName)
}

func (r *InMemoryPolicyGroupMemberRepository) AddMember(
	groupName string,
	identifier model.Identifier,
) error {
	// validate identifier/group name
	r.mu.Lock()
	defer r.mu.Unlock()

	r.members[identifier] = groupName

	return nil
}

func (r *InMemoryPolicyGroupMemberRepository) RemoveMember(
	identifier model.Identifier,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.members[identifier]
	if !ok {
		return ErrPolicyGroupMemberNotFound
	}

	delete(r.members, identifier)

	return nil
}
