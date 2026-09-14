```text
 ░▒▓███████▓▒░▒▓████████▓▒░▒▓█▓▒░░▒▓██████▓▒░░▒▓████████▓▒░▒▓███████▓▒░  
░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░ 
░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░ 
 ░▒▓██████▓▒░░▒▓██████▓▒░ ░▒▓█▓▒░▒▓█▓▒▒▓███▓▒░▒▓██████▓▒░ ░▒▓█▓▒░░▒▓█▓▒░ 
       ░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░ 
       ░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░ 
░▒▓███████▓▒░░▒▓████████▓▒░▒▓█▓▒░░▒▓██████▓▒░░▒▓████████▓▒░▒▓█▓▒░░▒▓█▓▒░ 

A lightweight, configurable rate-limiting library for Go.
```       

[![Go Reference](https://pkg.go.dev/badge/github.com/JhayceeCodes/seigen.svg)](https://pkg.go.dev/github.com/JhayceeCodes/seigen)
[![License](https://img.shields.io/github/license/JhayceeCodes/seigen)](LICENSE)                                                                 


## Features
- **Multiple rate-limiting algorithms**
  - Token Bucket
  - Leaky Bucket
  - Fixed Window
  - Sliding Window Log
  - Sliding Window Counter

- **Flexible policy management**
  - Individual policies
  - Policy groups
  - Individual policy precedence over group policies

- **Application-defined identifiers**
  - Implement `IdentifierResolver` to rate-limit by API key, user ID, IP address, tenant, or any other application-defined identifier.

- **Persistent policy storage**
  - PostgreSQL-backed policy and policy-group persistence
  - Pluggable repository interfaces

- **Concurrent-safe runtime limiting**
  - Thread-safe limiter implementations
  - Independent runtime limiter state per identifier

## Installation

```bash
go get github.com/JhayceeCodes/seigen
```

## Quick Start
Seigen separates rate-limit configuration from the runtime state used to enforce it.

The following example creates a rate limit for requests identified by an API key.

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/JhayceeCodes/seigen/limiter"
	"github.com/JhayceeCodes/seigen/model"
	"github.com/JhayceeCodes/seigen/service"
	"github.com/JhayceeCodes/seigen/store"
)

type APIKeyResolver struct{}

func (APIKeyResolver) Resolve(req *http.Request) (model.Identifier, error) {
	return model.Identifier(req.Header.Get("X-API-Key")), nil
}

func main() {
	// Store policy definitions in memory.
	policyStore := store.NewInMemoryPolicyRepository()

	// Create a rate limit policy.
	policy := model.Policy{
		Identifier: "client-123",
		Limiter: model.LimiterConfig{
			Algorithm: model.TokenBucket,
			Config: model.TokenBucketConfig{
				Capacity:       10,
				RefillInterval: time.Minute,
				RefillAmount:   10,
			},
		},
	}

	if err := policyStore.Set(policy); err != nil {
		panic(err)
	}

	// Create the runtime limiter manager.
	manager := limiter.NewManager()

	// Create the rate-limit service.
	resolver := APIKeyResolver{}

	rateLimitService := service.NewRateLimitService(
		resolver,
		policyStore,
		nil,
		manager,
	)

	// Evaluate an incoming request.
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "client-123")

	result, err := rateLimitService.Evaluate(req)
	if err != nil {
		panic(err)
	}

	fmt.Println("Allowed:", result.Allowed)
	fmt.Println("Remaining:", result.Remaining)
}
```

The resolver is application-defined, so you can extract an identifier from an API key, user ID, IP address, tenant ID, or any other request attribute.

For a persistent setup using PostgreSQL, see the [Integration Guide](#integration-guide).

## Core Concepts
Seigen is built around four core concepts:

- **Identifiers** — who or what is being rate-limited
- **Policies** — how an individual identifier is limited
- **Policy Groups** — shared rate-limit configuration for multiple identifiers
- **Runtime State** — the in-memory state used by the limiter to enforce the policy

### Identifiers
An identifier represents the entity being rate-limited.

```go
type Identifier string
```

An identifier can represent anything meaningful to your application, such as:
- API keys
- User IDs
- IP addresses
- Tenant IDs
- Client IDs

Seigen does not decide how an identifier is extracted from a request. Instead, your application implements `IdentifierResolver`:
```go
type IdentifierResolver interface {
	Resolve(*http.Request) (model.Identifier, error)
}
```


### Policies
A policy associates an identifier with a rate-limiting configuration.

```go
type Policy struct {
	Identifier model.Identifier
	Limiter    model.LimiterConfig
}
```
For example:
```go
model.Policy{
	Identifier: "client-123",
	Limiter: model.LimiterConfig{
		Algorithm: model.TokenBucket,
		Config: model.TokenBucketConfig{
			Capacity:       100,
			RefillInterval: time.Minute,
			RefillAmount:   100,
		},
	},
}
```
This means `client-123` gets its own rate-limiting configuration.

Policies are stored through a `PolicyRepository`, allowing the storage implementation to be changed without changing the rate-limiting service.

### Policy Groups
A policy group allows multiple identifiers to use the same rate-limiting configuration.

```go
type PolicyGroup struct {
	Name    string
	Limiter model.LimiterConfig
}
```
Identifiers are then assigned to the group:
```text
premium
├── client-123
├── client-456
└── client-789
```
For example, if `client-123` and `client-456` belong to the same group, they both use the group's rate-limit configuration, but each identifier has its own runtime limiter state. Consuming capacity for `client-123` does not consume capacity for `client-456`.

Policy groups are useful when multiple clients should follow the same rate-limit configuration without creating a separate policy definition for every client. A common use case is tier-based systems.

### Runtime State
A rate-limit policy describes **how** an identifier should be limited. The limiter's runtime state tracks **what has happened so far**.

For example, a token bucket needs to track information such as the current number of available tokens and when tokens were last refilled.

Seigen maintains runtime limiter instances in memory through the `limiter.Manager`.

When a request is evaluated, the manager returns the runtime limiter associated with the identifier:
```text
Request
   │
   ▼
Identifier Resolver
   │
   ▼
Identifier
   │
   ▼
Policy / Policy Group
   │
   ▼
Limiter Manager
   │
   ▼
Runtime Limiter
   │
   ▼
Allow / Reject
```
> Each identifier gets its own runtime limiter instance.

## Rate-Limiting Algorithms

Seigen supports five rate-limiting algorithms. Each algorithm uses a different strategy for controlling request frequency and has different trade-offs.

### Choosing an Algorithm

| Algorithm | Best suited for | Key characteristic |
|---|---|---|
| **Token Bucket** | APIs, general-purpose rate limiting | Allows controlled bursts while maintaining an average rate |
| **Leaky Bucket** | Traffic smoothing, protecting downstream services | Processes requests at a controlled rate and limits bursts |
| **Fixed Window** | Simple API quotas and basic rate limits | Simple and efficient, but can allow bursts at window boundaries |
| **Sliding Window Log** | Precise request-rate enforcement | Tracks individual requests for accurate rolling-window limits |
| **Sliding Window Counter** | High-throughput APIs where memory efficiency matters | Approximates a sliding window with lower memory usage |

### Token Bucket

Token Bucket maintains a bucket of tokens with a fixed capacity. Each request consumes a token, while tokens are replenished at a configured rate.

Because tokens can accumulate when the client is idle, the client can make a short burst of requests when tokens are available.

### Leaky Bucket

Leaky Bucket controls the rate at which requests are processed by allowing them to enter a bucket and "leaking" them at a configured interval.

Unlike Token Bucket, which naturally allows bursts when tokens have accumulated, Leaky Bucket is primarily useful for smoothing traffic.

### Fixed Window

Fixed Window divides time into fixed intervals. A request is allowed while the number of requests within the current window remains below the configured limit.

For example, a policy might allow 100 requests per minute.

The algorithm is simple and inexpensive, but requests near the boundary between two windows can produce a short burst that exceeds the intended rolling rate.

### Sliding Window Log

Sliding Window Log tracks the timestamp of each request and evaluates requests within a rolling time window.

Because individual request timestamps are tracked, this provides precise enforcement of the configured limit.

The trade-off is higher memory usage compared with simpler window-based approaches.


### Sliding Window Counter

Sliding Window Counter approximates a rolling window by combining request counts from the current and previous fixed windows.

It provides a balance between the simplicity of Fixed Window and the accuracy of Sliding Window Log while requiring significantly less memory than storing individual request timestamps.


## Integration Guide 

### Implementing an Identifier Resolver
Seigen does not extract identifiers from HTTP requests automatically. Your application defines how a request maps to a `model.Identifier` by implementing `IdentifierResolver`.

For example, an application that identifies clients using an API key can implement:

```go
type APIKeyResolver struct{}

func (APIKeyResolver) Resolve(req *http.Request) (model.Identifier, error) {
	apiKey := req.Header.Get("X-API-Key")

	if apiKey == "" {
		return "", errors.New("missing API key")
	}

	return model.Identifier(apiKey), nil
}
```
The same interface can be used with other identification strategies, such as authenticated user IDs, IP addresses, tenant IDs, or client IDs.

### Configuring PostgreSQL
Seigen provides a PostgreSQL repository for persistent policy storage.

Create a PostgreSQL connection using your application's database URL:

```go
db, err := database.NewPostgres(os.Getenv("DATABASE_URL"))
if err != nil {
	log.Fatal(err)
}
defer db.Close()
```

Run migrations:
```go
if err := migrations.Migrate(db); err != nil {
	log.Fatal(err)
}
```
Then create the policy repositories
```go
policyStore := store.NewPostgresPolicyRepository(db)
groupStore := store.NewPostgresPolicyGroupRepository(db)
groupMemberStore := store.NewPostgresPolicyGroupMemberRepository(db)
```

### Creating Policies
Create an individual policy by associating an identifier with a limiter configuration.

#### Token Bucket

```go
policy := model.Policy{
	Identifier: "client-123",
	Limiter: model.LimiterConfig{
		Algorithm: model.TokenBucket,
		Config: model.TokenBucketConfig{
			Capacity:       100,
			RefillInterval: time.Minute,
			RefillAmount:   100,
		},
	},
}
```

#### Leaky Bucket
```go
policy := model.Policy{
	Identifier: "client-123",
	Limiter: model.LimiterConfig{
		Algorithm: model.LeakyBucket,
		Config: model.LeakyBucketConfig{
			Capacity:    100,
			LeakInterval: time.Second,
		},
	},
}
```

#### Fixed Window
```go
policy := model.Policy{
	Identifier: "client-123",
	Limiter: model.LimiterConfig{
		Algorithm: model.FixedWindow,
		Config: model.WindowConfig{
			Limit: 100,
			Window: time.Minute,
		},
	},
}
```

#### Sliding Window Log
```go
policy := model.Policy{
	Identifier: "client-123",
	Limiter: model.LimiterConfig{
		Algorithm: model.SlidingWindowLog,
		Config: model.WindowConfig{
			Limit: 100,
			Window: time.Minute,
		},
	},
}
```

#### Sliding Window Counter
```go
policy := model.Policy{
	Identifier: "client-123",
	Limiter: model.LimiterConfig{
		Algorithm: model.SlidingWindowCounter,
		Config: model.WindowConfig{
			Limit: 100,
			Window: time.Minute,
		},
	},
}
```

Each policy can then be stored using the configured PolicyRepository:
```go
if err := policyStore.Set(policy); err != nil {
	log.Fatal(err)
}
```
The policy definition can be retrieved, updated, or deleted through the `PolicyRepository`.

### Creating Policy  Groups
Use a policy group when multiple identifiers should follow the same rate-limit configuration.

```go
group := model.PolicyGroup{
	Name: "premium",
	Limiter: model.LimiterConfig{
		Algorithm: model.TokenBucket,
		Config: model.TokenBucketConfig{
			Capacity:       500,
			RefillInterval: time.Minute,
			RefillAmount:   500,
		},
	},
}

if err := groupStore.Set(group); err != nil {
	log.Fatal(err)
}

// Assign identifiers to the group
if err := groupMemberStore.AddMember("premium", "client-123"); err != nil {
	log.Fatal(err)
}

if err := groupMemberStore.AddMember("premium", "client-456"); err != nil {
	log.Fatal(err)
}
```

### Evaluating Requests
Create a `RateLimitService` with the resolver, repositories, and limiter manager:

```go
resolver := APIKeyResolver{}
manager := limiter.NewManager()

rateLimitService := service.NewRateLimitService(
	resolver,
	policyStore,
	groupMemberStore,
	manager,
)

// Evaluate incoming request
result, err := rateLimitService.Evaluate(req)
if err != nil {
	// Handle the error according to your application's policy.
	log.Fatal(err)
}

if !result.Allowed {
	// Reject the request.
}
```
The returned `RateLimitResult` contains the limiter result and the configured limit: 
```go
fmt.Println("Allowed:", result.Allowed)
fmt.Println("Remaining:", result.Remaining)
fmt.Println("Limit:", result.Limit)
```

#### Applying Multiple Limits

An application can enforce multiple independent rate limits for the same request.

For example, you may want to limit both an API key and the originating IP address:

```go
apiKeyService := service.NewRateLimitService(
	apiKeyResolver,
	policyStore,
	groupMemberStore,
	manager,
)

ipService := service.NewRateLimitService(
	ipResolver,
	policyStore,
	groupMemberStore,
	manager,
)
```
Both services can then be evaluated for the same request:
```go
apiKeyResult, err := apiKeyService.Evaluate(req)
if err != nil {
	// Handle the error.
}

ipResult, err := ipService.Evaluate(req)
if err != nil {
	// Handle the error.
}

if !apiKeyResult.Allowed || !ipResult.Allowed {
	// Reject the request.
}
```
This allows applications to combine limits such as:
- API key + IP address
- User ID + IP address
- Tenant + API key
Any other combination of application-defined identifiers


### HTTP Middleware 
Seigen does not provide HTTP middleware. This keeps request handling and HTTP response behavior under the application's control.

A simple middleware can evaluate a request before passing it to the next handler:

```go
func RateLimitMiddleware(
	rateLimitService *service.RateLimitService,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		result, err := rateLimitService.Evaluate(req)
		if err != nil {
			http.Error(w, "rate limit evaluation failed", http.StatusInternalServerError)
			return
		}

		if !result.Allowed {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, req)
	})
}
```

## Policy Precedence

An identifier can have an individual policy, belong to a policy group, or both.

When both exist, the individual policy takes precedence.

```text
Individual Policy
       │
       ├── exists ──> Use individual policy
       │
       └── absent
              │
              ▼
        Policy Group
              │
              ├── member ──> Use group policy
              │
              └── not a member ──> No policy
```

## Testing
Seigen's test suite covers the limiter implementations, policy storage, service behavior, and concurrency safety.

Run the complete test suite with:

```bash
go test ./...
```
Run the tests with the race detector:
```bash
go test -race ./...
```
You can also run static analysis with:
```bash
go vet ./...
```

## Benchmarks
Benchmarks were run on the following system:

- **OS:** Debian 13 (Trixie)
- **Architecture:** linux/amd64
- **CPU:** Intel(R) Core(TM) i5-7200U CPU @ 2.50GHz
- **Go:** 1.27

The benchmarks cover the individual rate-limiting algorithms, limiter manager, rate-limit service, and concurrent request evaluation.

Run the benchmarks with:

```bash
go test ./benchmarks -bench=. -benchmem
```
**Results**
| Benchmark                                       | ns/op | B/op | allocs/op |
| ----------------------------------------------- | ----: | ---: | --------: |
| TokenBucketAllow                                | 65.39 |    0 |         0 |
| FixedWindowAllow                                | 68.96 |    0 |         0 |
| LeakyBucketAllow                                | 194.2 |  132 |         0 |
| SlidingWindowLogAllow                           | 213.2 |  139 |         0 |
| SlidingWindowCounterAllow                       | 151.4 |    0 |         0 |
| ManagerGetOrCreate                              | 49.98 |    0 |         0 |
| RateLimitServiceEvaluate                        | 299.9 |    0 |         0 |
| TokenBucketAllowParallel                        | 101.9 |    0 |         0 |
| RateLimitServiceEvaluateParallel_SameIdentifier | 307.6 |    0 |         0 |


## License
Seigen is licensed under the MIT License.

See the [License](LICENSE) file for the complete license text.