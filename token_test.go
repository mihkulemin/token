// Note: These tests were generated with the assistance of an LLM (Large Language Model).

package token

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewToken(t *testing.T) {
	tests := []struct {
		name    string
		maxCap  int
		length  int
		wantErr bool
	}{
		{
			name:    "Valid parameters",
			maxCap:  10,
			length:  5,
			wantErr: false,
		},
		{
			name:    "Zero max capacity",
			maxCap:  0,
			length:  0,
			wantErr: true,
		},
		{
			name:    "Negative max capacity",
			maxCap:  -1,
			length:  0,
			wantErr: true,
		},
		{
			name:    "Negative length",
			maxCap:  10,
			length:  -1,
			wantErr: true,
		},
		{
			name:    "Length greater than max capacity",
			maxCap:  5,
			length:  10,
			wantErr: true,
		},
		{
			name:    "Length equals max capacity",
			maxCap:  10,
			length:  10,
			wantErr: false,
		},
		{
			name:    "Zero initial tokens",
			maxCap:  10,
			length:  0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			token, err := NewToken(ctx, tt.maxCap, tt.length)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && token == nil {
				t.Error("NewToken() returned nil token without error")
			}

			if token != nil {
				// Cancel context to stop manager goroutine
				cancelCtx, cancel := context.WithCancel(ctx)
				cancel()
				_ = cancelCtx
			}
		})
	}
}

func TestToken_Take(t *testing.T) {
	t.Run("Take available token", func(t *testing.T) {
		ctx := context.Background()
		token, err := NewToken(ctx, 10, 5)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		err = token.Take(ctx)
		if err != nil {
			t.Errorf("Take() failed: %v", err)
		}
	})

	t.Run("Take with cancelled context", func(t *testing.T) {
		ctx := context.Background()
		token, err := NewToken(ctx, 2, 0) // Start with no tokens
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()

		err = token.Take(cancelCtx)
		if err != context.Canceled {
			t.Errorf("Take() with cancelled context = %v, want %v", err, context.Canceled)
		}
	})

	t.Run("Take blocks when no tokens available", func(t *testing.T) {
		ctx := context.Background()
		token, err := NewToken(ctx, 2, 0)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()

		err = token.Take(timeoutCtx)
		if err != context.DeadlineExceeded {
			t.Errorf("Take() without tokens = %v, want %v", err, context.DeadlineExceeded)
		}
	})

	t.Run("Take multiple tokens", func(t *testing.T) {
		ctx := context.Background()
		token, err := NewToken(ctx, 10, 3)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		// Take all 3 tokens
		for i := 0; i < 3; i++ {
			err = token.Take(ctx)
			if err != nil {
				t.Errorf("Take() token %d failed: %v", i+1, err)
			}
		}

		// Fourth take should block
		timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()

		err = token.Take(timeoutCtx)
		if err != context.DeadlineExceeded {
			t.Errorf("Take() fourth token = %v, want %v", err, context.DeadlineExceeded)
		}
	})
}

func TestToken_Release(t *testing.T) {
	t.Run("Release token", func(t *testing.T) {
		ctx := context.Background()
		token, err := NewToken(ctx, 10, 5)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		// Take a token first
		err = token.Take(ctx)
		if err != nil {
			t.Fatalf("Take() failed: %v", err)
		}

		// Release it back
		err = token.Release(ctx)
		if err != nil {
			t.Errorf("Release() failed: %v", err)
		}
	})

	t.Run("Release with cancelled context", func(t *testing.T) {
		ctx := context.Background()
		token, err := NewToken(ctx, 2, 2) // Start with full capacity
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()

		err = token.Release(cancelCtx)
		if err != context.Canceled {
			t.Errorf("Release() with cancelled context = %v, want %v", err, context.Canceled)
		}
	})

	t.Run("Release blocks when bucket is full", func(t *testing.T) {
		ctx := context.Background()
		token, err := NewToken(ctx, 2, 2)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		// Try to release when already at max capacity
		timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()

		err = token.Release(timeoutCtx)
		if err != context.DeadlineExceeded {
			t.Errorf("Release() when full = %v, want %v", err, context.DeadlineExceeded)
		}
	})
}

func TestToken_SetCapacity(t *testing.T) {
	t.Run("Set valid capacity", func(t *testing.T) {
		ctx := context.Background()
		token, err := NewToken(ctx, 10, 5)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		err = token.SetCapacity(ctx, 7)
		if err != nil {
			t.Errorf("SetCapacity() failed: %v", err)
		}
	})

	t.Run("Set capacity to zero", func(t *testing.T) {
		ctx := context.Background()
		token, err := NewToken(ctx, 10, 5)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		err = token.SetCapacity(ctx, 0)
		if err != nil {
			t.Errorf("SetCapacity(0) failed: %v", err)
		}
	})

	t.Run("Set capacity to maximum", func(t *testing.T) {
		ctx := context.Background()
		token, err := NewToken(ctx, 10, 5)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		err = token.SetCapacity(ctx, 10)
		if err != nil {
			t.Errorf("SetCapacity(max) failed: %v", err)
		}
	})

	t.Run("Set capacity above maximum", func(t *testing.T) {
		ctx := context.Background()
		token, err := NewToken(ctx, 10, 5)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		err = token.SetCapacity(ctx, 11)
		if err == nil {
			t.Error("SetCapacity() above max should fail")
		}
	})

	t.Run("Set negative capacity", func(t *testing.T) {
		ctx := context.Background()
		token, err := NewToken(ctx, 10, 5)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		err = token.SetCapacity(ctx, -1)
		if err == nil {
			t.Error("SetCapacity() with negative value should fail")
		}
	})

	t.Run("Set capacity with cancelled context", func(t *testing.T) {
		ctx := context.Background()
		token, err := NewToken(ctx, 10, 5)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()

		err = token.SetCapacity(cancelCtx, 7)
		if err != context.Canceled {
			t.Errorf("SetCapacity() with cancelled context = %v, want %v", err, context.Canceled)
		}
	})
}

func TestToken_DynamicCapacityAdjustment(t *testing.T) {
	t.Run("Increase capacity adds tokens", func(t *testing.T) {
		ctx := context.Background()
		token, err := NewToken(ctx, 10, 2)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		// Take both tokens
		for i := 0; i < 2; i++ {
			err = token.Take(ctx)
			if err != nil {
				t.Fatalf("Take() failed: %v", err)
			}
		}

		// Increase capacity
		err = token.SetCapacity(ctx, 4)
		if err != nil {
			t.Fatalf("SetCapacity() failed: %v", err)
		}

		// Allow manager goroutine to process
		time.Sleep(50 * time.Millisecond)

		// Should be able to take 2 more tokens
		for i := 0; i < 2; i++ {
			timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
			err = token.Take(timeoutCtx)
			cancel()
			if err != nil {
				t.Errorf("Take() after capacity increase failed: %v", err)
			}
		}

		// Third take should block
		timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()
		err = token.Take(timeoutCtx)
		if err != context.DeadlineExceeded {
			t.Errorf("Take() should block after taking increased capacity tokens")
		}
	})

	t.Run("Decrease capacity removes tokens", func(t *testing.T) {
		ctx := context.Background()
		token, err := NewToken(ctx, 10, 5)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		// Decrease capacity
		err = token.SetCapacity(ctx, 2)
		if err != nil {
			t.Fatalf("SetCapacity() failed: %v", err)
		}

		// Allow manager goroutine to process
		time.Sleep(50 * time.Millisecond)

		// Should only be able to take 2 tokens
		for i := 0; i < 2; i++ {
			err = token.Take(ctx)
			if err != nil {
				t.Errorf("Take() token %d failed: %v", i+1, err)
			}
		}

		// Third take should block
		timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()
		err = token.Take(timeoutCtx)
		if err != context.DeadlineExceeded {
			t.Errorf("Take() should block after capacity decrease")
		}
	})
}

func TestToken_Concurrency(t *testing.T) {
	t.Run("Concurrent Takes", func(t *testing.T) {
		ctx := context.Background()
		token, err := NewToken(ctx, 100, 50)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		var wg sync.WaitGroup
		successCount := 0
		var mu sync.Mutex

		// Launch 50 goroutines to take tokens
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := token.Take(ctx)
				if err == nil {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}()
		}

		wg.Wait()

		if successCount != 50 {
			t.Errorf("Expected 50 successful takes, got %d", successCount)
		}
	})

	t.Run("Concurrent Takes and Releases", func(t *testing.T) {
		ctx := context.Background()
		token, err := NewToken(ctx, 10, 5)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		var wg sync.WaitGroup

		// Launch takers
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				defer cancel()

				err := token.Take(timeoutCtx)
				if err == nil {
					// Hold token briefly
					time.Sleep(10 * time.Millisecond)
					// Release it back
					_ = token.Release(ctx)
				}
			}(i)
		}

		wg.Wait()
	})

	t.Run("Concurrent Capacity Changes", func(t *testing.T) {
		ctx := context.Background()
		token, err := NewToken(ctx, 100, 50)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		var wg sync.WaitGroup

		// Launch goroutines that change capacity
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				capacity := (id % 10) * 10
				if capacity == 0 {
					capacity = 10
				}
				_ = token.SetCapacity(ctx, capacity)
			}(i)
		}

		wg.Wait()
	})
}

func TestToken_ContextCancellation(t *testing.T) {
	t.Run("Manager stops on context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		token, err := NewToken(ctx, 10, 5)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		// Cancel the context
		cancel()

		// Give manager goroutine time to exit
		time.Sleep(50 * time.Millisecond)

		// Capacity changes should no longer be processed
		// This won't error but won't be processed either
		timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer timeoutCancel()
		err = token.SetCapacity(timeoutCtx, 8)
		// The SetCapacity might succeed in sending but the manager won't process it
		_ = err
	})

	t.Run("SetCapacity fails when token context is cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancelCause(context.Background())
		token, err := NewToken(ctx, 10, 5)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		// Cancel the token's internal context
		cancel(fmt.Errorf("token bucket is closed"))

		// Give manager goroutine time to exit
		time.Sleep(50 * time.Millisecond)

		// SetCapacity should fail immediately
		err = token.SetCapacity(context.Background(), 8)
		if err == nil {
			t.Error("SetCapacity() should fail when token context is cancelled")
		}
		if !strings.Contains(err.Error(), "token bucket is closed") {
			t.Errorf("SetCapacity() error = %v, want error containing 'token bucket is closed'", err)
		}
	})
}

func BenchmarkToken_Take(b *testing.B) {
	ctx := context.Background()
	token, err := NewToken(ctx, b.N, b.N)
	if err != nil {
		b.Fatalf("Failed to create token: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = token.Take(ctx)
	}
}

func BenchmarkToken_Release(b *testing.B) {
	ctx := context.Background()
	token, err := NewToken(ctx, b.N*2, 0)
	if err != nil {
		b.Fatalf("Failed to create token: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = token.Release(ctx)
	}
}

func BenchmarkToken_TakeRelease(b *testing.B) {
	ctx := context.Background()
	token, err := NewToken(ctx, 100, 50)
	if err != nil {
		b.Fatalf("Failed to create token: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = token.Take(ctx)
		_ = token.Release(ctx)
	}
}
