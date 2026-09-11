package identifier

import (
	"net/http"

	"github.com/JhayceeCodes/seigen/model"
)

// IdentifierResolver extracts the identifier used to select a rate-limiting
// policy from an incoming HTTP request.
//
// Applications can implement this interface to rate-limit requests by API key,
// user ID, IP address, tenant ID, or any other application-defined identifier.
type IdentifierResolver interface {
	Resolve(r *http.Request) (model.Identifier, error)
}
