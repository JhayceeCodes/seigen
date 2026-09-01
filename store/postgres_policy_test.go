package store_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/JhayceeCodes/seigen/model"
	"github.com/JhayceeCodes/seigen/store"
)

func TestPostgresPolicyRepository_SetAndGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	repository := store.NewPostgresPolicyRepository(db)

	policy := model.Policy{
		Identifier: "test-key",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.TokenBucketConfig{
				Capacity:       5,
				RefillInterval: time.Second,
				RefillAmount:   1,
			},
		},
	}

	// We expect Set() to execute an INSERT.
	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO seigen_policies (identifier, limiter_config)
		VALUES ($1, $2)
		ON CONFLICT (identifier)
		DO UPDATE SET limiter_config = EXCLUDED.limiter_config
	`)).
		WithArgs(policy.Identifier, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repository.Set(policy); err != nil {
		t.Fatalf("failed to set policy: %v", err)
	}

	// We expect Get() to execute a SELECT.
	configJSON := []byte(`{
		"algorithm": "token_bucket",
		"configuration": {
			"capacity": 5,
			"refill_interval": 1000000000,
			"refill_amount": 1
		}
	}`)

	rows := sqlmock.NewRows([]string{
		"identifier",
		"limiter_config",
	}).AddRow(
		policy.Identifier,
		configJSON,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT identifier, limiter_config
		FROM seigen_policies
		WHERE identifier = $1
	`)).
		WithArgs(policy.Identifier).
		WillReturnRows(rows)

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

	// Make sure every expectation was actually satisfied.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet database expectations: %v", err)
	}
}
