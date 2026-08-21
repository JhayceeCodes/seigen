package identifier

import (
	"net/http"

	"github.com/JhayceeCodes/seigen/internal/model"
)

// resolves a request to a rate-limiting identifier.
type IdentifierResolver interface {
	Resolve(r *http.Request) (model.Identifier, error)
}
