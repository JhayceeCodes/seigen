package store

import (
	"errors"

	"github.com/JhayceeCodes/seigen/internal/model"
)

type Store interface {
	Get(identifier model.Identifier) (model.Policy, error)
	Set(policy model.Policy) error
	Delete(identifier model.Identifier) error
}

var ErrPolicyNotFound = errors.New("policy not found")