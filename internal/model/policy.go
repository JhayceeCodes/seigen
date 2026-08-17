package model

import "errors"

type Policy struct {
	Identifier Identifier
	Limiter    LimiterConfig
}

func (p Policy) Validate() error {
	if p.Identifier == "" {
		return errors.New("identifier cannot be empty")
	}

	return p.Limiter.Validate()
}
