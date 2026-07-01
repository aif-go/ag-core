package agbatch

import (
	"context"
	"math"
	"time"
)

// RetryPolicy decides whether a failed item should be retried and after what delay.
type RetryPolicy interface {
	// ShouldRetry returns (shouldRetry, delayBeforeRetry).
	// attempt is 1-based — the first retry is attempt 1.
	ShouldRetry(attempt int, err error) (bool, time.Duration)
}

// SkipPolicy decides whether a failed item should be skipped (ignored)
// instead of failing the step.
type SkipPolicy interface {
	ShouldSkip(err error, skipCount int) bool
}

// --- Built-in Retry Policies ---

// MaxAttempts returns a policy that retries up to maxAttempts times
// with a fixed delay between attempts.
func MaxAttempts(maxAttempts int, delay time.Duration) RetryPolicy {
	return &simpleRetryPolicy{maxAttempts: maxAttempts, delay: delay}
}

type simpleRetryPolicy struct {
	maxAttempts int
	delay       time.Duration
}

func (p *simpleRetryPolicy) ShouldRetry(attempt int, _ error) (bool, time.Duration) {
	if attempt < p.maxAttempts {
		return true, p.delay
	}
	return false, 0
}

// ExponentialBackoff returns a policy that retries with exponentially increasing delays:
// delay * 2^(attempt-1), capped at maxDelay.
func ExponentialBackoff(maxAttempts int, initialDelay, maxDelay time.Duration) RetryPolicy {
	return &exponentialRetryPolicy{
		maxAttempts:  maxAttempts,
		initialDelay: initialDelay,
		maxDelay:     maxDelay,
	}
}

type exponentialRetryPolicy struct {
	maxAttempts  int
	initialDelay time.Duration
	maxDelay     time.Duration
}

func (p *exponentialRetryPolicy) ShouldRetry(attempt int, _ error) (bool, time.Duration) {
	if attempt >= p.maxAttempts {
		return false, 0
	}
	delay := time.Duration(float64(p.initialDelay) * math.Pow(2, float64(attempt-1)))
	if delay > p.maxDelay {
		delay = p.maxDelay
	}
	return true, delay
}

// NoRetry is a policy that never retries.
func NoRetry() RetryPolicy { return &noRetryPolicy{} }

type noRetryPolicy struct{}

func (p *noRetryPolicy) ShouldRetry(_ int, _ error) (bool, time.Duration) { return false, 0 }

// RetryableError returns a policy that wraps another policy but only retries
// when the error satisfies the predicate.
func RetryableError(delegate RetryPolicy, predicate func(error) bool) RetryPolicy {
	return &conditionalRetryPolicy{delegate: delegate, predicate: predicate}
}

type conditionalRetryPolicy struct {
	delegate  RetryPolicy
	predicate func(error) bool
}

func (p *conditionalRetryPolicy) ShouldRetry(attempt int, err error) (bool, time.Duration) {
	if !p.predicate(err) {
		return false, 0
	}
	return p.delegate.ShouldRetry(attempt, err)
}

// --- Built-in Skip Policies ---

// SkipLimit returns a policy that skips up to limit items.
func SkipLimit(limit int) SkipPolicy {
	return &skipLimitPolicy{limit: limit}
}

type skipLimitPolicy struct{ limit int }

func (p *skipLimitPolicy) ShouldSkip(_ error, skipCount int) bool {
	return skipCount < p.limit
}

// NeverSkip returns a policy that never skips — every error fails the step.
func NeverSkip() SkipPolicy { return &neverSkipPolicy{} }

type neverSkipPolicy struct{}

func (p *neverSkipPolicy) ShouldSkip(_ error, _ int) bool { return false }

// SkipOnError returns a policy that skips items when the error satisfies the predicate,
// up to the given limit.
func SkipOnError(predicate func(error) bool, limit int) SkipPolicy {
	return &conditionalSkipPolicy{predicate: predicate, limit: limit}
}

type conditionalSkipPolicy struct {
	predicate func(error) bool
	limit     int
}

func (p *conditionalSkipPolicy) ShouldSkip(err error, skipCount int) bool {
	return p.predicate(err) && skipCount < p.limit
}

// AlwaysSkip returns a policy that always skips on any error.
// Use with caution — only for non-critical batch steps.
func AlwaysSkip() SkipPolicy { return &alwaysSkipPolicy{} }

type alwaysSkipPolicy struct{}

func (p *alwaysSkipPolicy) ShouldSkip(_ error, _ int) bool { return true }

// Ensure context import is used.
var _ = context.Background
