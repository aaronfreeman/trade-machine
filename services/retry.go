package services

import (
	"context"
	"fmt"
	"log"
	"time"
)

type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

var DefaultRetryConfig = RetryConfig{
	MaxRetries:     3,
	InitialBackoff: 100 * time.Millisecond,
	MaxBackoff:     5 * time.Second,
}

func WithRetry(ctx context.Context, config RetryConfig, fn func() error) error {
	var lastErr error
	backoff := config.InitialBackoff

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(backoff):
			}

			backoff *= 2
			if backoff > config.MaxBackoff {
				backoff = config.MaxBackoff
			}
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err
		if attempt < config.MaxRetries {
			log.Printf("Retry attempt %d/%d failed: %v", attempt+1, config.MaxRetries, err)
		}
	}

	return fmt.Errorf("failed after %d retries: %w", config.MaxRetries, lastErr)
}
