package identifier

import (
	"errors"
	"net/http"
	"strings"

	"github.com/JhayceeCodes/seigen/internal/model"
)

type APIKeyResolver struct{}

func NewAPIKeyResolver() *APIKeyResolver {
	return &APIKeyResolver{}
}

func (r *APIKeyResolver) Resolve(req *http.Request) (model.Identifier, error) {
	authHeader := req.Header.Get("Authorization")

	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	parts := strings.SplitN(authHeader, " ", 2)

	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("authorization header format must be 'Bearer <token>'")
	}

	key := strings.TrimSpace(parts[1])

	if key == "" {
		return "", errors.New("api key is required")
	}

	if strings.ContainsAny(key, " \t\r\n") {
		return "", errors.New("api key cannot contain whitespace")
	}

	return model.Identifier(key), nil
}
