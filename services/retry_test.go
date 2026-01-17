package services

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWithRetry_Success(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
	}

	callCount := 0
	err := WithRetry(ctx, config, func() error {
		callCount++
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestWithRetry_EventualSuccess(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
	}

	callCount := 0
	err := WithRetry(ctx, config, func() error {
		callCount++
		if callCount < 3 {
			return errors.New("temporary error")
		}
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestWithRetry_AllFail(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxRetries:     2,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
	}

	callCount := 0
	expectedErr := errors.New("persistent error")
	err := WithRetry(ctx, config, func() error {
		callCount++
		return expectedErr
	})

	if err == nil {
		t.Error("expected error, got nil")
	}

	if callCount != 3 {
		t.Errorf("expected 3 calls (initial + 2 retries), got %d", callCount)
	}
}

func TestWithRetry_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 50 * time.Millisecond,
		MaxBackoff:     200 * time.Millisecond,
	}

	callCount := 0
	err := WithRetry(ctx, config, func() error {
		callCount++
		if callCount == 2 {
			cancel()
		}
		return errors.New("error")
	})

	if err == nil {
		t.Error("expected error, got nil")
	}

	if callCount > 3 {
		t.Errorf("expected at most 3 calls before cancellation, got %d", callCount)
	}
}

func TestWithRetry_ExponentialBackoff(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     50 * time.Millisecond,
	}

	startTime := time.Now()
	callCount := 0
	WithRetry(ctx, config, func() error {
		callCount++
		return errors.New("error")
	})
	duration := time.Since(startTime)

	expectedMinDuration := 10*time.Millisecond + 20*time.Millisecond + 40*time.Millisecond

	if duration < expectedMinDuration {
		t.Errorf("expected duration >= %v, got %v", expectedMinDuration, duration)
	}
}

func TestWithRetry_MaxBackoff(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     30 * time.Millisecond,
	}

	startTime := time.Now()
	WithRetry(ctx, config, func() error {
		return errors.New("error")
	})
	duration := time.Since(startTime)

	expectedMaxDuration := 30*time.Millisecond*5 + 100*time.Millisecond

	if duration > expectedMaxDuration {
		t.Errorf("backoff seems too long, expected < %v, got %v", expectedMaxDuration, duration)
	}
}
