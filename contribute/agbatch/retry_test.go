package agbatch

import (
	"errors"
	"testing"
	"time"
)

func TestMaxAttempts(t *testing.T) {
	p := MaxAttempts(3, 10*time.Millisecond)
	err := errors.New("fail")

	tests := []struct {
		attempt      int
		wantRetry    bool
		wantDelayMin time.Duration
	}{
		{1, true, 10 * time.Millisecond},
		{2, true, 10 * time.Millisecond},
		{3, false, 0},
		{4, false, 0},
	}

	for _, tt := range tests {
		shouldRetry, delay := p.ShouldRetry(tt.attempt, err)
		if shouldRetry != tt.wantRetry {
			t.Errorf("attempt %d: got retry=%v, want %v", tt.attempt, shouldRetry, tt.wantRetry)
		}
		if shouldRetry && delay != tt.wantDelayMin {
			t.Errorf("attempt %d: got delay=%v, want %v", tt.attempt, delay, tt.wantDelayMin)
		}
	}
}

func TestExponentialBackoff(t *testing.T) {
	p := ExponentialBackoff(4, 10*time.Millisecond, 100*time.Millisecond)

	tests := []struct {
		attempt   int
		wantRetry bool
		wantDelay time.Duration
	}{
		{1, true, 10 * time.Millisecond}, // 10 * 2^0
		{2, true, 20 * time.Millisecond}, // 10 * 2^1
		{3, true, 40 * time.Millisecond}, // 10 * 2^2
		{4, false, 0},                    // max attempts reached
	}

	for _, tt := range tests {
		shouldRetry, delay := p.ShouldRetry(tt.attempt, errors.New("fail"))
		if shouldRetry != tt.wantRetry {
			t.Errorf("attempt %d: got retry=%v, want %v", tt.attempt, shouldRetry, tt.wantRetry)
		}
		if shouldRetry && delay != tt.wantDelay {
			t.Errorf("attempt %d: got delay=%v, want %v", tt.attempt, delay, tt.wantDelay)
		}
	}
}

func TestExponentialBackoffCapped(t *testing.T) {
	p := ExponentialBackoff(5, 100*time.Millisecond, 200*time.Millisecond)

	_, delay := p.ShouldRetry(3, errors.New("fail")) // 100 * 2^2 = 400 capped to 200
	if delay != 200*time.Millisecond {
		t.Errorf("delay should be capped at 200ms, got %v", delay)
	}
}

func TestNoRetry(t *testing.T) {
	p := NoRetry()
	shouldRetry, _ := p.ShouldRetry(1, errors.New("fail"))
	if shouldRetry {
		t.Error("NoRetry should not retry")
	}
}

func TestRetryableError(t *testing.T) {
	p := RetryableError(MaxAttempts(3, time.Millisecond), func(err error) bool {
		return err.Error() == "retryable"
	})

	shouldRetry, _ := p.ShouldRetry(1, errors.New("retryable"))
	if !shouldRetry {
		t.Error("should retry retryable errors")
	}

	shouldRetry, _ = p.ShouldRetry(1, errors.New("fatal"))
	if shouldRetry {
		t.Error("should not retry non-retryable errors")
	}
}

func TestSkipLimit(t *testing.T) {
	p := SkipLimit(5)
	err := errors.New("fail")

	tests := []struct {
		skipCount int
		wantSkip  bool
	}{
		{0, true},
		{1, true},
		{4, true},
		{5, false},
		{10, false},
	}

	for _, tt := range tests {
		got := p.ShouldSkip(err, tt.skipCount)
		if got != tt.wantSkip {
			t.Errorf("skipCount=%d: got skip=%v, want %v", tt.skipCount, got, tt.wantSkip)
		}
	}
}

func TestNeverSkip(t *testing.T) {
	p := NeverSkip()
	if p.ShouldSkip(errors.New("fail"), 0) {
		t.Error("NeverSkip should not skip")
	}
}

func TestSkipOnError(t *testing.T) {
	p := SkipOnError(func(err error) bool {
		return err.Error() == "skip_me"
	}, 3)

	if !p.ShouldSkip(errors.New("skip_me"), 0) {
		t.Error("should skip matching error when under limit")
	}
	if p.ShouldSkip(errors.New("other"), 0) {
		t.Error("should not skip non-matching error")
	}
	if p.ShouldSkip(errors.New("skip_me"), 3) {
		t.Error("should not skip when over limit")
	}
}

func TestAlwaysSkip(t *testing.T) {
	p := AlwaysSkip()
	if !p.ShouldSkip(errors.New("anything"), 999) {
		t.Error("AlwaysSkip should always skip")
	}
}
