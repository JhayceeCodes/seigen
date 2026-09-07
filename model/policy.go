package model

import "errors"

type Policy struct {
	Identifier Identifier
	Limiter    LimiterConfig
}

type PolicyGroup struct {
	Name    string
	Limiter LimiterConfig
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

func (p Policy) Equal(other Policy) bool {
	return p.Identifier == other.Identifier &&
		p.Limiter.Equal(other.Limiter)
}

func (c LimiterConfig) Equal(other LimiterConfig) bool {
	return c.Algorithm == other.Algorithm && c.Config == other.Config
}
