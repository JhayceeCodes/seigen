package identifier_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JhayceeCodes/seigen/identifier"
)

func TestAPIKeyResolverReturnsIdentifier(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer sgn-123")

	resolver := identifier.NewAPIKeyResolver()

	id, err := resolver.Resolve(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if id != "sgn-123" {
		t.Fatalf("expected api key to be extracted successfully, got %v", id)
	}
}

func TestAPIKeyResolverIsCaseInsensitive(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "bearer sgn-123")

	resolver := identifier.NewAPIKeyResolver()

	id, err := resolver.Resolve(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if id != "sgn-123" {
		t.Fatalf("expected api key to be extracted successfully, got %v", id)
	}
}

func TestAPIKeyResolverRejectsMissingHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	resolver := identifier.NewAPIKeyResolver()

	_, err := resolver.Resolve(req)
	if err == nil {
		t.Fatal("expected error for missing authorization header")
	}
}

func TestAPIKeyResolverRejectsInvalidScheme(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bear sgn-123")

	resolver := identifier.NewAPIKeyResolver()

	_, err := resolver.Resolve(req)
	if err == nil {
		t.Fatal("expected error for invalid scheme")
	}
}

func TestAPIKeyResolverRejectsMissingToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer  ")

	resolver := identifier.NewAPIKeyResolver()

	_, err := resolver.Resolve(req)
	if err == nil {
		t.Fatal("expected error for missing token")
	}
}

func TestAPIKeyResolverRejectsTokenWithSpaces(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer sgn-123 extra")

	resolver := identifier.NewAPIKeyResolver()

	_, err := resolver.Resolve(req)
	if err == nil {
		t.Fatal("expected error for API key containing whitespace")
	}
}
