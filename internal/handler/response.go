package handler

import "github.com/JhayceeCodes/rate-limiter-gateway/internal/model"

type APIKeyResponse struct {
	Status  string       `json:"status"`
	Message string       `json:"message"`
	Data    model.APIKey `json:"data"`
}
