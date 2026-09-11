package integration

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/JhayceeCodes/seigen/database"
	"github.com/JhayceeCodes/seigen/identifier"
	"github.com/JhayceeCodes/seigen/limiter"
	"github.com/JhayceeCodes/seigen/migrations"
	"github.com/JhayceeCodes/seigen/model"
	"github.com/JhayceeCodes/seigen/service"
	"github.com/JhayceeCodes/seigen/store"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("SEIGEN_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("SEIGEN_TEST_DATABASE_URL is not set")
	}

	db, err := database.NewPostgres(dsn)
	require.NoError(t, err)

	require.NoError(t, migrations.Migrate(db))

	t.Cleanup(func() {
		_, err := db.Exec(`
			TRUNCATE
				seigen_policy_group_members,
				seigen_policy_groups,
				seigen_policies
			RESTART IDENTITY CASCADE
		`)
		require.NoError(t, err)

		require.NoError(t, db.Close())
	})

	return db
}

func newRateLimitService(db *sql.DB) (
	*service.RateLimitService,
	store.PolicyRepository,
	store.PolicyGroupRepository,
	store.PolicyGroupMemberRepository,
) {
	policyStore := store.NewPostgresPolicyRepository(db)
	groupStore := store.NewPostgresPolicyGroupRepository(db)
	groupMemberStore := store.NewPostgresPolicyGroupMemberRepository(db)

	manager := limiter.NewManager()
	resolver := identifier.NewAPIKeyResolver()

	rateLimitService := service.NewRateLimitService(
		resolver,
		policyStore,
		groupMemberStore,
		manager,
	)

	return rateLimitService, policyStore, groupStore, groupMemberStore
}

func newRequest(apiKey string) *http.Request {
	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	req.Header.Set("Authorization", "Bearer "+apiKey)

	return req
}

func tokenBucketConfig(capacity int) model.LimiterConfig {
	return model.LimiterConfig{
		Algorithm: model.TokenBucket,
		Config: model.TokenBucketConfig{
			Capacity:       capacity,
			RefillInterval: time.Minute,
			RefillAmount:   capacity,
		},
	}
}

func TestRateLimitService_GroupPolicyFallback(t *testing.T) {
	db := setupTestDB(t)

	rateLimitService, _, groupStore, groupMemberStore :=
		newRateLimitService(db)

	err := groupStore.Set(model.PolicyGroup{
		Name:    "premium",
		Limiter: tokenBucketConfig(5),
	})
	require.NoError(t, err)

	err = groupMemberStore.AddMember("premium", "key-1")
	require.NoError(t, err)

	result, err := rateLimitService.Evaluate(newRequest("key-1"))

	require.NoError(t, err)
	require.True(t, result.Allowed)
	require.Equal(t, 5, result.Limit)
	require.Equal(t, 4, result.Remaining)
}

func TestRateLimitService_IndividualPolicyTakesPrecedence(t *testing.T) {
	db := setupTestDB(t)

	rateLimitService, policyStore, groupStore, groupMemberStore :=
		newRateLimitService(db)

	// Group allows 5 requests.
	err := groupStore.Set(model.PolicyGroup{
		Name:    "premium",
		Limiter: tokenBucketConfig(5),
	})
	require.NoError(t, err)

	err = groupMemberStore.AddMember("premium", "key-1")
	require.NoError(t, err)

	// Individual policy only allows 2 requests.
	err = policyStore.Set(model.Policy{
		Identifier: "key-1",
		Limiter:    tokenBucketConfig(2),
	})
	require.NoError(t, err)

	// The individual policy must win over the group policy.
	result, err := rateLimitService.Evaluate(newRequest("key-1"))

	require.NoError(t, err)
	require.True(t, result.Allowed)
	require.Equal(t, 2, result.Limit)
	require.Equal(t, 1, result.Remaining)

	// Second request should consume the final token.
	result, err = rateLimitService.Evaluate(newRequest("key-1"))

	require.NoError(t, err)
	require.True(t, result.Allowed)
	require.Equal(t, 2, result.Limit)
	require.Equal(t, 0, result.Remaining)

	// Third request should be rejected.
	result, err = rateLimitService.Evaluate(newRequest("key-1"))

	require.NoError(t, err)
	require.False(t, result.Allowed)
	require.Equal(t, 2, result.Limit)
	require.Equal(t, 0, result.Remaining)
}

func TestRateLimitService_GroupMembersHaveIndependentRuntimeState(t *testing.T) {
	db := setupTestDB(t)

	rateLimitService, _, groupStore, groupMemberStore :=
		newRateLimitService(db)

	err := groupStore.Set(model.PolicyGroup{
		Name:    "premium",
		Limiter: tokenBucketConfig(2),
	})
	require.NoError(t, err)

	err = groupMemberStore.AddMember("premium", "key-1")
	require.NoError(t, err)

	err = groupMemberStore.AddMember("premium", "key-2")
	require.NoError(t, err)

	// Exhaust key-1.
	result, err := rateLimitService.Evaluate(newRequest("key-1"))
	require.NoError(t, err)
	require.True(t, result.Allowed)
	require.Equal(t, 1, result.Remaining)

	result, err = rateLimitService.Evaluate(newRequest("key-1"))
	require.NoError(t, err)
	require.True(t, result.Allowed)
	require.Equal(t, 0, result.Remaining)

	// key-1 should now be rate limited.
	result, err = rateLimitService.Evaluate(newRequest("key-1"))
	require.NoError(t, err)
	require.False(t, result.Allowed)
	require.Equal(t, 0, result.Remaining)

	// key-2 has its own runtime limiter and should still have
	// the full group allowance.
	result, err = rateLimitService.Evaluate(newRequest("key-2"))

	require.NoError(t, err)
	require.True(t, result.Allowed)
	require.Equal(t, 2, result.Limit)
	require.Equal(t, 1, result.Remaining)
}

func TestRateLimitService_UnknownIdentifier(t *testing.T) {
	db := setupTestDB(t)

	rateLimitService, _, _, _ := newRateLimitService(db)

	result, err := rateLimitService.Evaluate(
		newRequest("unknown-key"),
	)

	require.Error(t, err)
	require.False(t, result.Allowed)
	require.ErrorIs(t, err, store.ErrPolicyGroupMemberNotFound)
}

func TestRateLimitService_IndividualPolicyDoesNotRequireGroupMembership(t *testing.T) {
	db := setupTestDB(t)

	rateLimitService, policyStore, _, _ :=
		newRateLimitService(db)

	err := policyStore.Set(model.Policy{
		Identifier: "key-1",
		Limiter:    tokenBucketConfig(2),
	})
	require.NoError(t, err)

	// key-1 has no group membership, but its individual policy
	// should be sufficient.
	result, err := rateLimitService.Evaluate(newRequest("key-1"))

	require.NoError(t, err)
	require.True(t, result.Allowed)
	require.Equal(t, 2, result.Limit)
	require.Equal(t, 1, result.Remaining)
}

func TestRateLimitService_GroupMemberUsesPersistedConfiguration(t *testing.T) {
	db := setupTestDB(t)

	rateLimitService, _, groupStore, groupMemberStore :=
		newRateLimitService(db)

	err := groupStore.Set(model.PolicyGroup{
		Name:    "standard",
		Limiter: tokenBucketConfig(3),
	})
	require.NoError(t, err)

	err = groupMemberStore.AddMember("standard", "key-1")
	require.NoError(t, err)

	// The service must retrieve the configuration from PostgreSQL
	// and construct the runtime limiter from it.
	result, err := rateLimitService.Evaluate(newRequest("key-1"))

	require.NoError(t, err)
	require.True(t, result.Allowed)
	require.Equal(t, 3, result.Limit)
	require.Equal(t, 2, result.Remaining)
}

func TestRateLimitService_IndividualPolicyAndGroupUseSeparateRuntimeState(t *testing.T) {
	db := setupTestDB(t)

	rateLimitService, policyStore, groupStore, groupMemberStore :=
		newRateLimitService(db)

	err := groupStore.Set(model.PolicyGroup{
		Name:    "premium",
		Limiter: tokenBucketConfig(5),
	})
	require.NoError(t, err)

	err = groupMemberStore.AddMember("premium", "key-1")
	require.NoError(t, err)

	err = groupMemberStore.AddMember("premium", "key-2")
	require.NoError(t, err)

	// key-1 has an individual policy.
	err = policyStore.Set(model.Policy{
		Identifier: "key-1",
		Limiter:    tokenBucketConfig(2),
	})
	require.NoError(t, err)

	// key-1 uses its individual limiter.
	result, err := rateLimitService.Evaluate(newRequest("key-1"))
	require.NoError(t, err)
	require.Equal(t, 2, result.Limit)
	require.Equal(t, 1, result.Remaining)

	// key-2 uses the group limiter.
	result, err = rateLimitService.Evaluate(newRequest("key-2"))
	require.NoError(t, err)
	require.Equal(t, 5, result.Limit)
	require.Equal(t, 4, result.Remaining)
}
