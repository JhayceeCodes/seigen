package identifier_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JhayceeCodes/seigen/identifier"
)

func TestIPResolverParsesIPV4Correctly(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:54321"

	resolver := identifier.NewIPResolver()

	id, err := resolver.Resolve(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if id != "127.0.0.1" {
		t.Fatalf("expected IP %q, got %q", "127.0.0.1", id)
	}
}

func TestIPResolverParsesIPV6Correctly(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "[::1]:54321"

	resolver := identifier.NewIPResolver()

	id, err := resolver.Resolve(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if id != "::1" {
		t.Fatalf("expected IP %q, got %q", "::1", id)
	}
}

func TestIPResolverRejectsInvalidAddress(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "not-a-valid-address"

	resolver := identifier.NewIPResolver()

	_, err := resolver.Resolve(req)
	if err == nil {
		t.Fatal("expected error for invalid remote address")
	}
}

func TestIPResolverRejectsEmptyIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = ""

	resolver := identifier.NewIPResolver()

	_, err := resolver.Resolve(req)
	if err == nil {
		t.Fatal("expected error for empty remote address")
	}
}
