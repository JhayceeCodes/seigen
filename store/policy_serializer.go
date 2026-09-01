package store

import (
	"encoding/json"
	"fmt"

	"github.com/JhayceeCodes/seigen/model"
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

func deserializeLimiterConfig(data []byte) (model.LimiterConfig, error) {
	var payload limiterConfigJSON

	if err := json.Unmarshal(data, &payload); err != nil {
		return model.LimiterConfig{}, err
	}

	switch payload.Algorithm {
	case model.TokenBucket:
		var config model.TokenBucketConfig

		if err := json.Unmarshal(payload.Configuration, &config); err != nil {
			return model.LimiterConfig{}, err
		}

		return model.LimiterConfig{
			Algorithm: payload.Algorithm,
			Config:    config,
		}, nil

	case model.LeakyBucket:
		var config model.LeakyBucketConfig

		if err := json.Unmarshal(payload.Configuration, &config); err != nil {
			return model.LimiterConfig{}, err
		}

		return model.LimiterConfig{
			Algorithm: payload.Algorithm,
			Config:    config,
		}, nil

	case model.FixedWindow,
		model.SlidingWindowLog,
		model.SlidingWindowCounter:

		var config model.WindowConfig

		if err := json.Unmarshal(payload.Configuration, &config); err != nil {
			return model.LimiterConfig{}, err
		}

		return model.LimiterConfig{
			Algorithm: payload.Algorithm,
			Config:    config,
		}, nil

	default:
		return model.LimiterConfig{}, fmt.Errorf(
			"unsupported limiter algorithm: %q",
			payload.Algorithm,
		)
	}
}
