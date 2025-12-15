package overpass

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RetryConfig holds retry behavior configuration
type RetryConfig struct {
	MaxRetries        int           // Maximum retry attempts (default: 3)
	InitialBackoff    time.Duration // Initial backoff duration (default: 1s)
	MaxBackoff        time.Duration // Maximum backoff duration (default: 30s)
	BackoffMultiplier float64       // Backoff multiplier (default: 2.0)
	Jitter            bool          // Add randomization to prevent thundering herd (default: true)
}

// DefaultRetryConfig returns sensible defaults
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    time.Second,
		MaxBackoff:        30 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            true,
	}
}

// isRetryableStatus determines if HTTP status code warrants retry
func isRetryableStatus(statusCode int) bool {
	return statusCode == 429 || // Too Many Requests
		statusCode == 500 || // Internal Server Error
		statusCode == 502 || // Bad Gateway
		statusCode == 503 || // Service Unavailable
		statusCode == 504 // Gateway Timeout
}

// calculateBackoff computes next backoff duration
func calculateBackoff(attempt int, config RetryConfig) time.Duration {
	backoff := float64(config.InitialBackoff) * math.Pow(config.BackoffMultiplier, float64(attempt))

	if backoff > float64(config.MaxBackoff) {
		backoff = float64(config.MaxBackoff)
	}

	if config.Jitter {
		// Add up to 25% jitter to prevent thundering herd
		jitter := rand.Float64() * 0.25 * backoff
		backoff += jitter
	}

	return time.Duration(backoff)
}

// retryableHTTPPost wraps httpPost with retry logic
func (c *Client) retryableHTTPPost(ctx context.Context, query string) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		// Check context before attempting
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		body, err := c.httpPost(ctx, query)

		// Success - return immediately
		if err == nil {
			return body, nil
		}

		// Check if error is retryable
		var serverErr *ServerError
		isServerErr := errors.As(err, &serverErr)

		if !isServerErr || !isRetryableStatus(serverErr.StatusCode) {
			// Not retryable - return error immediately
			return nil, err
		}

		lastErr = err

		// Don't sleep after last attempt
		if attempt < c.retryConfig.MaxRetries {
			backoff := calculateBackoff(attempt, c.retryConfig)

			// Sleep with context awareness
			select {
			case <-time.After(backoff):
				// Continue to next attempt
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
