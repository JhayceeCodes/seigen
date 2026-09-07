package store_test

import (
	"database/sql"
	"errors"
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

func TestPostgresPolicyRepository_GetReturnsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	repository := store.NewPostgresPolicyRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT identifier, limiter_config
		FROM seigen_policies
		WHERE identifier = $1
	`)).
		WithArgs(model.Identifier("unknown-key")).
		WillReturnError(sql.ErrNoRows)

	_, err = repository.Get("unknown-key")

	if !errors.Is(err, store.ErrPolicyNotFound) {
		t.Fatalf("expected ErrPolicyNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet database expectations: %v", err)
	}
}

func TestPostgresPolicyRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	repository := store.NewPostgresPolicyRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM seigen_policies 
		WHERE identifier = $1
	`)).
		WithArgs(model.Identifier("test-key")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repository.Delete("test-key"); err != nil {
		t.Fatalf("failed to delete policy: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet database expectations: %v", err)
	}
}

func TestPostgresPolicyRepository_DeleteReturnsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	repository := store.NewPostgresPolicyRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM seigen_policies 
		WHERE identifier = $1
	`)).
		WithArgs(model.Identifier("unknown-key")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repository.Delete("unknown-key")

	if !errors.Is(err, store.ErrPolicyNotFound) {
		t.Fatalf("expected ErrPolicyNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet database expectations: %v", err)
	}
}

func TestPostgresPolicyRepository_SetRejectsInvalidPolicy(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	repository := store.NewPostgresPolicyRepository(db)

	policy := model.Policy{
		Identifier: "test-key",
		Limiter: model.LimiterConfig{
			Algorithm: model.FixedWindow,
			Config: model.TokenBucketConfig{
				Capacity:       5,
				RefillInterval: time.Second,
				RefillAmount:   1,
			},
		},
	}

	if err := repository.Set(policy); err == nil {
		t.Fatal("expected invalid policy to be rejected")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected database interaction: %v", err)
	}
}

func TestPostgresPolicyRepository_SetReplacesExistingPolicy(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	repository := store.NewPostgresPolicyRepository(db)

	policy := model.Policy{
		Identifier: "test-key",
		Limiter: model.LimiterConfig{
			Algorithm: model.FixedWindow,
			Config: model.WindowConfig{
				Limit:  10,
				Window: time.Minute,
			},
		},
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO seigen_policies (identifier, limiter_config)
		VALUES ($1, $2)
		ON CONFLICT (identifier)
		DO UPDATE SET limiter_config = EXCLUDED.limiter_config
	`)).
		WithArgs(policy.Identifier, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repository.Set(policy); err != nil {
		t.Fatalf("failed to set initial policy: %v", err)
	}

	updatedPolicy := model.Policy{
		Identifier: "test-key",
		Limiter: model.LimiterConfig{
			Algorithm: model.FixedWindow,
			Config: model.WindowConfig{
				Limit:  100,
				Window: time.Minute,
			},
		},
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO seigen_policies (identifier, limiter_config)
		VALUES ($1, $2)
		ON CONFLICT (identifier)
		DO UPDATE SET limiter_config = EXCLUDED.limiter_config
	`)).
		WithArgs(updatedPolicy.Identifier, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repository.Set(updatedPolicy); err != nil {
		t.Fatalf("failed to update policy: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet database expectations: %v", err)
	}
}


func TestPostgresPolicyGroupRepository_SetAndGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repository := store.NewPostgresPolicyGroupRepository(db)

	group := model.PolicyGroup{
		Name: "premium-users",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.TokenBucketConfig{
				Capacity:       100,
				RefillInterval: time.Second,
				RefillAmount:   10,
			},
		},
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO seigen_policy_groups (name, limiter_config)
		VALUES ($1, $2)
		ON CONFLICT (name)
		DO UPDATE SET limiter_config = EXCLUDED.limiter_config
	`)).
		WithArgs(group.Name, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repository.Set(group); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	configJSON := []byte(`{
		"algorithm": "token_bucket",
		"configuration": {
			"capacity": 100,
			"refill_interval": 1000000000,
			"refill_amount": 10
		}
	}`)

	rows := sqlmock.NewRows([]string{
		"name",
		"limiter_config",
	}).AddRow(
		group.Name,
		configJSON,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT name, limiter_config
		FROM seigen_policy_groups
		WHERE name = $1
	`)).
		WithArgs(group.Name).
		WillReturnRows(rows)

	got, err := repository.Get(group.Name)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.Name != group.Name {
		t.Errorf("Name = %q, want %q", got.Name, group.Name)
	}

	if got.Limiter.Algorithm != model.TokenBucket {
		t.Errorf(
			"Algorithm = %q, want %q",
			got.Limiter.Algorithm,
			model.TokenBucket,
		)
	}

	config, ok := got.Limiter.Config.(model.TokenBucketConfig)
	if !ok {
		t.Fatalf("Config type = %T, want model.TokenBucketConfig", got.Limiter.Config)
	}

	if config.Capacity != 100 {
		t.Errorf("Capacity = %d, want 100", config.Capacity)
	}

	if config.RefillInterval != time.Second {
		t.Errorf(
			"RefillInterval = %v, want %v",
			config.RefillInterval,
			time.Second,
		)
	}

	if config.RefillAmount != 10 {
		t.Errorf("RefillAmount = %d, want 10", config.RefillAmount)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPostgresPolicyGroupRepository_GetNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repository := store.NewPostgresPolicyGroupRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT name, limiter_config
		FROM seigen_policy_groups
		WHERE name = $1
	`)).
		WithArgs("unknown-group").
		WillReturnError(sql.ErrNoRows)

	_, err = repository.Get("unknown-group")
	if !errors.Is(err, store.ErrPolicyGroupNotFound) {
		t.Fatalf(
			"Get() error = %v, want %v",
			err,
			store.ErrPolicyGroupNotFound,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}


func TestPostgresPolicyGroupRepository_SetInvalidGroup(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repository := store.NewPostgresPolicyGroupRepository(db)

	group := model.PolicyGroup{
		Name: "",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.TokenBucketConfig{
				Capacity:       100,
				RefillInterval: time.Second,
				RefillAmount:   10,
			},
		},
	}

	err = repository.Set(group)
	if err == nil {
		t.Fatal("Set() expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected database interaction: %v", err)
	}
}

func TestPostgresPolicyGroupRepository_SetReplacesExisting(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repository := store.NewPostgresPolicyGroupRepository(db)

	group := model.PolicyGroup{
		Name: "premium-users",
		Limiter: model.LimiterConfig{
			Algorithm: model.FixedWindow,
			Config: model.WindowConfig{
				Limit:  100,
				Window: time.Minute,
			},
		},
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO seigen_policy_groups (name, limiter_config)
		VALUES ($1, $2)
		ON CONFLICT (name)
		DO UPDATE SET limiter_config = EXCLUDED.limiter_config
	`)).
		WithArgs(group.Name, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repository.Set(group); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}


func TestPostgresPolicyGroupMemberRepository_AddMember(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repository := store.NewPostgresPolicyGroupMemberRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO seigen_policy_group_members (group_id, identifier)
		SELECT id, $2
		FROM seigen_policy_groups
		WHERE name = $1
	`)).
		WithArgs("premium-users", model.Identifier("api-key-123")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repository.AddMember(
		"premium-users",
		model.Identifier("api-key-123"),
	)
	if err != nil {
		t.Fatalf("AddMember() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}


func TestPostgresPolicyGroupMemberRepository_AddMemberGroupNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repository := store.NewPostgresPolicyGroupMemberRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO seigen_policy_group_members (group_id, identifier)
		SELECT id, $2
		FROM seigen_policy_groups
		WHERE name = $1
	`)).
		WithArgs("unknown-group", model.Identifier("api-key-123")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repository.AddMember(
		"unknown-group",
		model.Identifier("api-key-123"),
	)

	if !errors.Is(err, store.ErrPolicyGroupNotFound) {
		t.Fatalf(
			"AddMember() error = %v, want %v",
			err,
			store.ErrPolicyGroupNotFound,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}


func TestPostgresPolicyGroupMemberRepository_AddMemberInvalidInput(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repository := store.NewPostgresPolicyGroupMemberRepository(db)

	tests := []struct {
		name       string
		groupName  string
		identifier model.Identifier
	}{
		{
			name:       "empty group name",
			groupName:  "",
			identifier: "api-key-123",
		},
		{
			name:       "empty identifier",
			groupName:  "premium-users",
			identifier: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repository.AddMember(tt.groupName, tt.identifier)
			if err == nil {
				t.Fatal("AddMember() expected error, got nil")
			}
		})
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected database interaction: %v", err)
	}
}



func TestPostgresPolicyGroupMemberRepository_GetGroup(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repository := store.NewPostgresPolicyGroupMemberRepository(db)

	configJSON := []byte(`{
		"algorithm": "token_bucket",
		"configuration": {
			"capacity": 50,
			"refill_interval": 1000000000,
			"refill_amount": 5
		}
	}`)

	rows := sqlmock.NewRows([]string{
		"name",
		"limiter_config",
	}).AddRow(
		"premium-users",
		configJSON,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT g.name, g.limiter_config
		FROM seigen_policy_group_members AS m
		JOIN seigen_policy_groups AS g
			ON g.id = m.group_id
		WHERE m.identifier = $1
	`)).
		WithArgs(model.Identifier("api-key-123")).
		WillReturnRows(rows)

	got, err := repository.GetGroup(model.Identifier("api-key-123"))
	if err != nil {
		t.Fatalf("GetGroup() error = %v", err)
	}

	if got.Name != "premium-users" {
		t.Errorf("Name = %q, want %q", got.Name, "premium-users")
	}

	if got.Limiter.Algorithm != model.TokenBucket {
		t.Errorf(
			"Algorithm = %q, want %q",
			got.Limiter.Algorithm,
			model.TokenBucket,
		)
	}

	config, ok := got.Limiter.Config.(model.TokenBucketConfig)
	if !ok {
		t.Fatalf(
			"Config type = %T, want model.TokenBucketConfig",
			got.Limiter.Config,
		)
	}

	if config.Capacity != 50 {
		t.Errorf("Capacity = %d, want 50", config.Capacity)
	}

	if config.RefillInterval != time.Second {
		t.Errorf(
			"RefillInterval = %v, want %v",
			config.RefillInterval,
			time.Second,
		)
	}

	if config.RefillAmount != 5 {
		t.Errorf("RefillAmount = %d, want 5", config.RefillAmount)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}


func TestPostgresPolicyGroupMemberRepository_GetGroupNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repository := store.NewPostgresPolicyGroupMemberRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT g.name, g.limiter_config
		FROM seigen_policy_group_members AS m
		JOIN seigen_policy_groups AS g
			ON g.id = m.group_id
		WHERE m.identifier = $1
	`)).
		WithArgs(model.Identifier("unknown-key")).
		WillReturnError(sql.ErrNoRows)

	_, err = repository.GetGroup(model.Identifier("unknown-key"))

	if !errors.Is(err, store.ErrPolicyGroupMemberNotFound) {
		t.Fatalf(
			"GetGroup() error = %v, want %v",
			err,
			store.ErrPolicyGroupMemberNotFound,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}


func TestPostgresPolicyGroupMemberRepository_RemoveMember(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repository := store.NewPostgresPolicyGroupMemberRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM seigen_policy_group_members
		WHERE identifier = $1
	`)).
		WithArgs(model.Identifier("api-key-123")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repository.RemoveMember(model.Identifier("api-key-123"))
	if err != nil {
		t.Fatalf("RemoveMember() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}


func TestPostgresPolicyGroupRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repository := store.NewPostgresPolicyGroupRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM seigen_policy_groups
		WHERE name = $1
	`)).
		WithArgs("premium-users").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repository.Delete("premium-users")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}



func TestPostgresPolicyGroupRepository_DeleteNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repository := store.NewPostgresPolicyGroupRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM seigen_policy_groups
		WHERE name = $1
	`)).
		WithArgs("unknown-group").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repository.Delete("unknown-group")

	if !errors.Is(err, store.ErrPolicyGroupNotFound) {
		t.Fatalf(
			"Delete() error = %v, want %v",
			err,
			store.ErrPolicyGroupNotFound,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}


