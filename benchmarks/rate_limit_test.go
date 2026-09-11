package benchmarks

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JhayceeCodes/seigen/limiter"
	"github.com/JhayceeCodes/seigen/model"
	"github.com/JhayceeCodes/seigen/service"
	"github.com/JhayceeCodes/seigen/store"
)

const benchmarkLimit = 1_000_000_000

// ------------------------------------------------------------
// Helpers
// ------------------------------------------------------------

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

func newBenchmarkRequest() *http.Request {
	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	req.Header.Set("X-Test-Identifier", "key-1")

	return req
}

type testResolver struct{}

func (testResolver) Resolve(req *http.Request) (model.Identifier, error) {
	return model.Identifier(req.Header.Get("X-Test-Identifier")), nil
}

func newBenchmarkService(b *testing.B) *service.RateLimitService {
	b.Helper()

	policyStore := store.NewInMemoryPolicyRepository()

	groupStore := store.NewInMemoryPolicyGroupRepository()

	groupMemberStore :=
		store.NewInMemoryPolicyGroupMemberRepository(groupStore)

	err := groupStore.Set(model.PolicyGroup{
		Name:    "premium",
		Limiter: tokenBucketConfig(benchmarkLimit),
	})
	if err != nil {
		b.Fatal(err)
	}

	err = groupMemberStore.AddMember(
		"premium",
		"key-1",
	)
	if err != nil {
		b.Fatal(err)
	}

	manager := limiter.NewManager()

	resolver := testResolver{}

	return service.NewRateLimitService(
		resolver,
		policyStore,
		groupMemberStore,
		manager,
	)
}

// ------------------------------------------------------------
// Limiter benchmarks
// ------------------------------------------------------------

func BenchmarkTokenBucketAllow(b *testing.B) {
	lim := limiter.NewTokenBucket(
		benchmarkLimit,
		time.Minute,
		benchmarkLimit,
	)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		lim.Allow()
	}
}

func BenchmarkFixedWindowAllow(b *testing.B) {
	lim := limiter.NewFixedWindow(
		benchmarkLimit,
		time.Minute,
	)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		lim.Allow()
	}
}

func BenchmarkLeakyBucketAllow(b *testing.B) {
	lim := limiter.NewLeakyBucket(
		benchmarkLimit,
		time.Minute,
	)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		lim.Allow()
	}
}

func BenchmarkSlidingWindowLogAllow(b *testing.B) {
	lim := limiter.NewSlidingWindowLog(
		benchmarkLimit,
		time.Minute,
	)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		lim.Allow()
	}
}

func BenchmarkSlidingWindowCounterAllow(b *testing.B) {
	lim := limiter.NewSlidingWindowCounter(
		benchmarkLimit,
		time.Minute,
	)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		lim.Allow()
	}
}

// ------------------------------------------------------------
// Manager benchmarks
// ------------------------------------------------------------

func BenchmarkManagerGetOrCreate(b *testing.B) {
	manager := limiter.NewManager()

	config := tokenBucketConfig(benchmarkLimit)

	// Warm up the manager so the benchmark measures the
	// existing-instance path rather than limiter creation.
	_, err := manager.GetOrCreate(
		"key-1",
		config,
	)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_, err := manager.GetOrCreate(
			"key-1",
			config,
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ------------------------------------------------------------
// Service benchmark
// ------------------------------------------------------------

func BenchmarkRateLimitServiceEvaluate(b *testing.B) {
	rateLimitService := newBenchmarkService(b)

	req := newBenchmarkRequest()

	// Warm up the manager so limiter construction isn't included
	// in the benchmark.
	_, err := rateLimitService.Evaluate(req)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_, err := rateLimitService.Evaluate(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ------------------------------------------------------------
// Concurrent benchmarks
// ------------------------------------------------------------

func BenchmarkTokenBucketAllowParallel(b *testing.B) {
	lim := limiter.NewTokenBucket(
		benchmarkLimit,
		time.Minute,
		benchmarkLimit,
	)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lim.Allow()
		}
	})
}

func BenchmarkRateLimitServiceEvaluateParallel_SameIdentifier(
	b *testing.B,
) {
	rateLimitService := newBenchmarkService(b)

	// Warm up the manager before starting parallel execution.
	req := newBenchmarkRequest()

	_, err := rateLimitService.Evaluate(req)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		// Each worker gets its own request object.
		req := newBenchmarkRequest()

		for pb.Next() {
			_, err := rateLimitService.Evaluate(req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
