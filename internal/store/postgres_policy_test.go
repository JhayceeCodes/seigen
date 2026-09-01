package store_test

import (
	"os"
	"testing"
	"time"

	"github.com/JhayceeCodes/seigen/internal/database"
	"github.com/JhayceeCodes/seigen/internal/model"
	"github.com/JhayceeCodes/seigen/internal/store"
	"github.com/joho/godotenv"
)

func TestPostgresPolicyRepository_SetAndGet(t *testing.T) {
	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf("failed to load .env: %v", err)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is not set")
	}

	db, err := database.NewPostgres(dsn)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	repository := store.NewPostgresPolicyRepository(db)

	policy := model.Policy{
		Identifier: "integration-test-key",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.TokenBucketConfig{
				Capacity:       5,
				RefillInterval: time.Second,
				RefillAmount:   1,
			},
		},
	}

	if err := repository.Set(policy); err != nil {
		t.Fatalf("failed to set policy %v", err)
	}

	got, err := repository.Get(policy.Identifier)
	if err != nil {
		t.Fatalf("failed to get policy: %v", err)
	}

	if got.Identifier != policy.Identifier {
		t.Errorf(
			"expected identifier %q, got %q",
			policy.Identifier,
			got.Identifier,
		)
	}

	if got.Limiter.Algorithm != policy.Limiter.Algorithm {
		t.Errorf(
			"expected algorithm %q, got %q",
			policy.Limiter.Algorithm,
			got.Limiter.Algorithm,
		)
	}

	config, ok := got.Limiter.Config.(model.TokenBucketConfig)
	if !ok {
		t.Fatalf(
			"expected TokenBucketConfig, got %T",
			got.Limiter.Config,
		)
	}

	expectedConfig := policy.Limiter.Config.(model.TokenBucketConfig)

	if config.Capacity != expectedConfig.Capacity {
		t.Errorf(
			"expected capacity %d, got %d",
			expectedConfig.Capacity,
			config.Capacity,
		)
	}

	if config.RefillInterval != expectedConfig.RefillInterval {
		t.Errorf(
			"expected refill interval %v, got %v",
			expectedConfig.RefillInterval,
			config.RefillInterval,
		)
	}

	if config.RefillAmount != expectedConfig.RefillAmount {
		t.Errorf(
			"expected refill amount %d, got %d",
			expectedConfig.RefillAmount,
			config.RefillAmount,
		)
	}
}
