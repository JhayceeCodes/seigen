package limiter

import (
	"errors"

	"github.com/JhayceeCodes/seigen/model"
)


// New creates a limiter using the algorithm and configuration specified by
// config.
//
// It returns an error if the algorithm is unsupported or the configuration
// is invalid.
func New(config model.LimiterConfig) (Limiter, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	switch config.Algorithm {
	case model.TokenBucket:
		cfg, ok := config.Config.(model.TokenBucketConfig)
		if !ok {
			return nil, errors.New("token bucket requires TokenBucketConfig")
		}
		return NewTokenBucket(
			cfg.Capacity,
			cfg.RefillInterval,
			cfg.RefillAmount,
		), nil

	case model.LeakyBucket:
		cfg, ok := config.Config.(model.LeakyBucketConfig)
		if !ok {
			return nil, errors.New("leaky bucket requires LeakyBucketConfig")
		}
		return NewLeakyBucket(
			cfg.Capacity,
			cfg.LeakInterval,
		), nil

	case model.FixedWindow:
		cfg, ok := config.Config.(model.WindowConfig)
		if !ok {
			return nil, errors.New("fixed window requires WindowConfig")
		}
		return NewFixedWindow(cfg.Limit, cfg.Window), nil

	case model.SlidingWindowLog:
		cfg, ok := config.Config.(model.WindowConfig)
		if !ok {
			return nil, errors.New("sliding window log requires WindowConfig")
		}
		return NewSlidingWindowLog(cfg.Limit, cfg.Window), nil

	case model.SlidingWindowCounter:
		cfg, ok := config.Config.(model.WindowConfig)
		if !ok {
			return nil, errors.New("sliding window counter requires WindowConfig")
		}
		return NewSlidingWindowCounter(cfg.Limit, cfg.Window), nil

	default:
		return nil, errors.New("unsupported limiter algorithm")
	}
}
