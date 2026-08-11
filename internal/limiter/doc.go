// Package limiter provides several in-memory rate limiting algorithms.
//
// Supported algorithms:
//
//   - Token Bucket
//   - Fixed Window
//   - Sliding Window Log
//	 - Sliding Window Counter
//   - Leaky Bucket
//
// The package is designed to be embedded into HTTP middleware and other
// request-processing pipelines.
package limiter
