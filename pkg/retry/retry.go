package retry

import (
	"context"
	"errors"
	"sync"
	"time"
)

var EmptyBackoff Backoff = BackoffFunc(func() (time.Duration, bool) {
	return 0, true
})

// TODO: add tests
// Backoff is an interface that backs off.
type Backoff interface {
	// Next returns the time duration to wait and whether to stop.
	Next() (next time.Duration, stop bool)
}

type BackoffFunc func() (time.Duration, bool)

// Next implements Backoff.
func (b BackoffFunc) Next() (time.Duration, bool) {
	return b()
}

func NewBaseBackoff() Backoff {
	return WithMaxRetries(3, NewLinearBackoff(1*time.Second, 2*time.Second))
}

func NewLinearBackoff(base time.Duration, step time.Duration) Backoff {
	duration := base

	return BackoffFunc(func() (time.Duration, bool) {
		prev := duration

		duration += step

		return prev, false
	})
}

func WithMaxRetries(max uint64, next Backoff) Backoff {
	var l sync.Mutex
	var attempt uint64

	return BackoffFunc(func() (time.Duration, bool) {
		l.Lock()
		defer l.Unlock()

		if attempt >= max {
			return 0, true
		}
		attempt++

		val, stop := next.Next()
		if stop {
			return 0, true
		}

		return val, false
	})
}

// RetryFunc is a function passed to retry.
type RetryFunc func(ctx context.Context) error
type RetryFuncWithData[T any] func(ctx context.Context) (T, error)

type retryableError struct {
	err error
}

// RetryableError marks an error as retryable.
func RetryableError(err error) error {
	if err == nil {
		return nil
	}
	return &retryableError{err}
}

// Unwrap implements error wrapping.
func (e *retryableError) Unwrap() error {
	return e.err
}

// Error returns the error string.
func (e *retryableError) Error() string {
	if e.err == nil {
		return "retryable: <nil>"
	}
	return "retryable: " + e.err.Error()
}

// Do wraps a function with a backoff to retry. The provided context is the same
// context passed to the RetryFunc.
func Do(ctx context.Context, b Backoff, f RetryFunc) error {
	for {
		// Return immediately if ctx is canceled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := f(ctx)
		if err == nil {
			return nil
		}

		// Not retryable
		var rerr *retryableError
		if !errors.As(err, &rerr) {
			return err
		}

		next, stop := b.Next()

		if stop {
			return rerr.Unwrap()
		}

		// ctx.Done() has priority, so we test it alone first
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		t := time.NewTimer(next)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
			continue
		}
	}
}

func DoWithData[T any](ctx context.Context, b Backoff, f RetryFuncWithData[T]) (T, error) {
	for {
		var emptyT T
		// Return immediately if ctx is canceled
		select {
		case <-ctx.Done():
			return emptyT, ctx.Err()
		default:
		}

		data, err := f(ctx)
		if err == nil {
			return data, nil
		}

		// Not retryable
		var rerr *retryableError
		if !errors.As(err, &rerr) {
			return data, err
		}

		next, stop := b.Next()

		if stop {
			return data, rerr.Unwrap()
		}

		// ctx.Done() has priority, so we test it alone first
		select {
		case <-ctx.Done():
			return data, ctx.Err()
		default:
		}

		t := time.NewTimer(next)
		select {
		case <-ctx.Done():
			t.Stop()
			return data, ctx.Err()
		case <-t.C:
			continue
		}
	}
}
