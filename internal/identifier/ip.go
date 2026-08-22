package identifier

import (
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/JhayceeCodes/seigen/internal/model"
)

type IPResolver struct{}

func NewIPResolver() *IPResolver {
	return &IPResolver{}
}

func (r *IPResolver) Resolve(req *http.Request) (model.Identifier, error) {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return "", fmt.Errorf("parse remote address: %w", err)
	}

	if ip == "" {
		return "", errors.New("remote address is empty")
	}

	return model.Identifier(ip), nil
}
