package store

import (
	"encoding/json"

	"github.com/JhayceeCodes/seigen/internal/model"
)

type limiterConfigJSON struct {
	Algorithm     model.Algorithm `json:"algorithm"`
	Configuration json.RawMessage `json:"configuration"`
}

func serializeLimiterConfig(
	config model.LimiterConfig,
) ([]byte, error) {
	configuration, err := json.Marshal(config.Config)
	if err != nil {
		return nil, err
	}

	payload := limiterConfigJSON{
		Algorithm:     config.Algorithm,
		Configuration: configuration,
	}

	return json.Marshal(payload)
}
