package model

import "errors"

// Policy associates an identifier with a rate-limiting configuration.
type Policy struct {
	Identifier Identifier
	Limiter    LimiterConfig
}

// PolicyGroup defines a shared rate-limiting configuration for multiple
// identifiers. Each member maintains independent runtime limiter state.
type PolicyGroup struct {
	Name    string
	Limiter LimiterConfig
}

// PolicyGroupMember associates an identifier with a policy group.
type PolicyGroupMember struct {
	GroupID    int64
	Identifier Identifier
}

func (p Policy) Validate() error {
	if p.Identifier == "" {
		return errors.New("identifier cannot be empty")
	}

	return p.Limiter.Validate()
}

func (g PolicyGroup) Validate() error {
	if g.Name == "" {
		return errors.New("group name cannot be empty")
	}

	return g.Limiter.Validate()
}

func (m PolicyGroupMember) Validate() error {
	if m.GroupID <= 0 {
		return errors.New("group ID must be greater than zero")
	}

	if m.Identifier == "" {
		return errors.New("identifier cannot be empty")
	}

	return nil
}

func (p Policy) Equal(other Policy) bool {
	return p.Identifier == other.Identifier &&
		p.Limiter.Equal(other.Limiter)
}

func (c LimiterConfig) Equal(other LimiterConfig) bool {
	return c.Algorithm == other.Algorithm && c.Config == other.Config
}
