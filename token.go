// Package token provides a dynamic token bucket implementation for rate limiting
// and resource management. It allows for runtime adjustment of token capacity
// while maintaining thread-safe operations through channels.
package token

import (
	"context"
	"fmt"
)

// Token represents a token bucket with dynamic capacity adjustment.
// It uses channels for thread-safe token management and supports
// context-based cancellation for all operations.
type Token struct {
	ctx context.Context
	buf chan struct{}
	// Maximum number of possible token
	maxCapacity int
	// Channel for capacity changes
	cap chan int
	// Actual number of tokens available
	length int
}

// NewToken creates a new Token bucket with the specified maximum capacity and initial length.
// The context is used for lifecycle management of the internal goroutine.
//
// Parameters:
//   - ctx: Context for managing the token bucket lifecycle
//   - maxCap: Maximum capacity of the token bucket (must be > 0)
//   - len: Initial number of tokens (must be >= 0 and <= maxCap)
//
// Returns an error if maxCap <= 0, len < 0, or len > maxCap.
func NewToken(ctx context.Context, maxCap, len int) (*Token, error) {
	if maxCap <= 0 || len < 0 || maxCap < len {
		return nil, fmt.Errorf("incorrect max capacity (%d) and/or length (%d)", maxCap, len)
	}

	t := &Token{
		ctx:         ctx,
		buf:         make(chan struct{}, maxCap),
		maxCapacity: maxCap,
		cap:         make(chan int),
		length:      len,
	}

	for i := 0; i < t.length; i++ {
		t.buf <- struct{}{}
	}

	go t.manager()

	return t, nil
}

// Take acquires a token from the bucket, blocking until one is available
// or the context is cancelled.
//
// Returns:
//   - nil if a token was successfully acquired
//   - context error if the operation was cancelled
func (t *Token) Take(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-t.ctx.Done():
		return context.Cause(t.ctx)
	case <-t.buf:
		return nil
	}
}

// Release returns a token to the bucket, blocking if the bucket is full
// until space is available or the context is cancelled.
//
// Returns:
//   - nil if the token was successfully released
//   - context error if the operation was cancelled
func (t *Token) Release(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-t.ctx.Done():
		return context.Cause(t.ctx)
	case t.buf <- struct{}{}:
		return nil

	}
}

// SetCapacity dynamically adjusts the number of available tokens in the bucket.
// The new capacity must be between 0 and the maximum capacity set during initialization.
// The adjustment is handled asynchronously by the internal manager goroutine.
//
// Parameters:
//   - ctx: Context for the operation
//   - c: New capacity (must be between 0 and maxCapacity)
//
// Returns an error if the capacity is out of bounds or the context is cancelled.
func (t *Token) SetCapacity(ctx context.Context, c int) error {
	if c > t.maxCapacity || c < 0 {
		return fmt.Errorf("capacity (%d) should be between 0 and maximum (%d)", c, t.maxCapacity)
	}

	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-t.ctx.Done():
		return context.Cause(t.ctx)
	case t.cap <- c:
	}
	return nil
}

// manager is an internal goroutine that handles dynamic capacity adjustments.
// It continuously monitors for capacity change requests and adjusts the number
// of tokens in the bucket accordingly by either adding or removing tokens.
func (t *Token) manager() {
	var in chan<- struct{}
	var out <-chan struct{}
	target := t.length

	for {
		select {
		case <-t.ctx.Done():
			return
		case in <- struct{}{}:
			t.length++
		case <-out:
			t.length--
		case target = <-t.cap:
		}

		if target > t.length {
			in = t.buf
			out = nil
		} else if target < t.length {
			in = nil
			out = t.buf
		} else {
			in = nil
			out = nil
		}
	}
}
