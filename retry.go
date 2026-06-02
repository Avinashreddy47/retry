// Package retry provides a simple, context-aware retry utility with
// configurable exponential backoff.
//
// Basic usage:
//
//	err := retry.Do(ctx, retry.DefaultConfig(), func() error {
//	    return callUnreliableAPI()
//	})
package retry

import (
	"context"
	"errors"
	"math"
	"time"
)

// Config controls retry behavior.
type Config struct {
	// MaxAttempts is the total number of attempts (first call + retries).
	// Must be >= 1.
	MaxAttempts int

	// Delay is the wait time before the second attempt.
	Delay time.Duration

	// MaxDelay caps the delay after multiplier is applied.
	MaxDelay time.Duration

	// Multiplier grows the delay on each subsequent attempt.
	// 1.0 means constant delay; 2.0 means doubling (exponential backoff).
	Multiplier float64
}

// DefaultConfig returns a Config with sensible production defaults:
// 3 attempts, 100ms initial delay, 10s cap, exponential backoff.
func DefaultConfig() Config {
	return Config{
		MaxAttempts: 3,
		Delay:       100 * time.Millisecond,
		MaxDelay:    10 * time.Second,
		Multiplier:  2.0,
	}
}

// Permanent wraps an error to signal that it must not be retried.
// Use this when you know a failure is definitive (e.g. 400 Bad Request).
func Permanent(err error) error {
	return &permanentError{err: err}
}

// IsPermanent reports whether err was wrapped with [Permanent].
func IsPermanent(err error) bool {
	var p *permanentError
	return errors.As(err, &p)
}

type permanentError struct{ err error }

func (p *permanentError) Error() string { return p.err.Error() }
func (p *permanentError) Unwrap() error { return p.err }

// Do calls fn up to cfg.MaxAttempts times. It stops early if:
//   - fn returns nil
//   - fn returns an error wrapped with [Permanent]
//   - ctx is cancelled or times out
//
// The delay between attempts grows by cfg.Multiplier each round,
// capped at cfg.MaxDelay. Do returns the last error from fn, or
// ctx.Err() if the context expired while waiting.
func Do(ctx context.Context, cfg Config, fn func() error) error {
	if cfg.MaxAttempts < 1 {
		cfg.MaxAttempts = 1
	}

	var err error
	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err = fn()
		if err == nil {
			return nil
		}
		if IsPermanent(err) {
			return errors.Unwrap(err)
		}

		if attempt < cfg.MaxAttempts-1 {
			delay := nextDelay(cfg.Delay, cfg.Multiplier, cfg.MaxDelay, attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}
	return err
}

func nextDelay(base time.Duration, multiplier float64, max time.Duration, attempt int) time.Duration {
	d := float64(base) * math.Pow(multiplier, float64(attempt))
	if d > float64(max) {
		return max
	}
	return time.Duration(d)
}
