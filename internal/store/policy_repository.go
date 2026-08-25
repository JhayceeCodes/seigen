package store

import (
	"github.com/JhayceeCodes/seigen/internal/model"
)

type PolicyRepository interface {
	Set(policy model.Policy) error
	Get(identifier model.Identifier) (model.Policy, error)
	Delete(identifier model.Identifier) error
}

