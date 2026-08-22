package identifier_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JhayceeCodes/seigen/internal/identifier"
	"github.com/JhayceeCodes/seigen/internal/model"
)

type mockResolver struct {
	id model.Identifier
}

func (m *mockResolver) Resolve(r *http.Request) (model.Identifier, error) {
	return m.id, nil
}

func TestCustomResolverReturnsIdentifier(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	resolver := &mockResolver{
		id: "user:123",
	}

	var identifierResolver identifier.IdentifierResolver = resolver

	got, err := identifierResolver.Resolve(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got != "user:123" {
		t.Fatalf("expected identifier %q, got %q", "user:123", got)
	}
}

type userIDResolver struct{}

func (r *userIDResolver) Resolve(req *http.Request) (model.Identifier, error) {
	userID := req.Header.Get("X-User-ID")

	return model.Identifier(userID), nil
}

func TestCustomResolverCanUseRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-ID", "user:123")

	resolver := &userIDResolver{}

	var identifierResolver identifier.IdentifierResolver = resolver

	got, err := identifierResolver.Resolve(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got != "user:123" {
		t.Fatalf("expected identifier %q, got %q", "user:123", got)
	}
}
