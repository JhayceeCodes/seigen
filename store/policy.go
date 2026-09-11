package store

import (
	"github.com/JhayceeCodes/seigen/model"
)

// PolicyRepository persists and retrieves individual rate-limiting policies.
type PolicyRepository interface {
	Set(policy model.Policy) error
	Get(identifier model.Identifier) (model.Policy, error)
	Delete(identifier model.Identifier) error
}

// PolicyGroupRepository persists and retrieves policy group definitions.
type PolicyGroupRepository interface {
	Set(policyGroup model.PolicyGroup) error
	Get(name string) (model.PolicyGroup, error)
	Delete(name string) error
}

// PolicyGroupMemberRepository manages associations between identifiers and
// policy groups.
type PolicyGroupMemberRepository interface {
	AddMember(groupName string, identifier model.Identifier) error
	RemoveMember(identifier model.Identifier) error
	GetGroup(identifier model.Identifier) (model.PolicyGroup, error)
}
