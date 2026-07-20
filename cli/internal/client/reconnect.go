package client

import (
	"context"
	"math"
	"time"
)

// Reconnect manages exponential backoff for connection retries.
type Reconnect struct {
	MaxRetries int
	attempt    int
	baseDelay  time.Duration
	maxDelay   time.Duration
}

// NewReconnect creates a reconnect manager with a max retry count (0 = unlimited).
func NewReconnect(maxRetries int) *Reconnect {
	return &Reconnect{
		MaxRetries: maxRetries,
		baseDelay:  1 * time.Second,
		maxDelay:   30 * time.Second,
	}
}

// Sleep waits for the appropriate backoff duration. Returns false if max retries exceeded.
// Use SleepContext to support cancellation.
func (r *Reconnect) Sleep() bool {
	if r.MaxRetries > 0 && r.attempt >= r.MaxRetries {
		return false
	}
	delay := time.Duration(math.Min(
		float64(r.baseDelay)*math.Pow(2, float64(r.attempt)),
		float64(r.maxDelay),
	))
	time.Sleep(delay)
	r.attempt++
	return true
}

// SleepContext waits for the backoff or until ctx is cancelled.
// Returns false if max retries exceeded or ctx is cancelled.
func (r *Reconnect) SleepContext(ctx context.Context) bool {
	ok, _ := r.sleepContext(ctx, nil)
	return ok
}

// sleepContext waits for the backoff, context cancellation, or an explicit
// wake-up. The second return value reports whether wake interrupted the wait.
func (r *Reconnect) sleepContext(ctx context.Context, wake <-chan struct{}) (bool, bool) {
	if r.MaxRetries > 0 && r.attempt >= r.MaxRetries {
		return false, false
	}
	delay := time.Duration(math.Min(
		float64(r.baseDelay)*math.Pow(2, float64(r.attempt)),
		float64(r.maxDelay),
	))
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false, false
	case <-wake:
		return true, true
	case <-timer.C:
	}
	r.attempt++
	return true, false
}

// Reset clears the attempt counter after a successful connection.
func (r *Reconnect) Reset() {
	r.attempt = 0
}
