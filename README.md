# Token Bucket

*Note: This README was generated with the assistance of an LLM (Large Language Model).*

A Go package that provides a dynamic token bucket implementation for rate limiting and resource management.

## Features

- **Dynamic Capacity Adjustment**: Adjust the number of available tokens at runtime without restarting
- **Context-Based Operations**: All operations support context cancellation for proper timeout and cleanup handling
- **Thread-Safe**: Uses Go channels for safe concurrent access
- **Non-Blocking Operations**: All operations can be cancelled via context
- **Efficient Resource Management**: Internal goroutine manages token adjustments asynchronously

## Installation

```bash
go get github.com/mihkulemin/token
```

## Usage

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/mihkulemin/token"
)

func main() {
    ctx := context.Background()
    
    // Create a token bucket with max capacity of 10 and initial 5 tokens
    bucket, err := token.NewToken(ctx, 10, 5)
    if err != nil {
        panic(err)
    }
    
    // Take a token
    if err := bucket.Take(ctx); err != nil {
        fmt.Println("Failed to take token:", err)
    }
    
    // Do some work...
    
    // Release the token back
    if err := bucket.Release(ctx); err != nil {
        fmt.Println("Failed to release token:", err)
    }
}
```

### Dynamic Capacity Adjustment

```go
// Adjust capacity at runtime
ctx := context.Background()
bucket, _ := token.NewToken(ctx, 100, 50)

// Reduce available tokens to 25
err := bucket.SetCapacity(ctx, 25)

// Increase available tokens to 75
err = bucket.SetCapacity(ctx, 75)
```

### With Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Will timeout after 5 seconds if no token is available
if err := bucket.Take(ctx); err != nil {
    if err == context.DeadlineExceeded {
        fmt.Println("Timeout waiting for token")
    }
}
```

## API Reference

### `NewToken(ctx context.Context, maxCap, len int) (*Token, error)`

Creates a new token bucket with specified maximum capacity and initial token count.

**Parameters:**
- `ctx`: Context for managing the token bucket lifecycle
- `maxCap`: Maximum capacity of the token bucket (must be > 0)
- `len`: Initial number of tokens (must be >= 0 and <= maxCap)

**Returns:**
- `*Token`: The created token bucket
- `error`: Error if parameters are invalid

### `Take(ctx context.Context) error`

Acquires a token from the bucket. Blocks until a token is available or context is cancelled.

**Parameters:**
- `ctx`: Context for the operation

**Returns:**
- `nil` if successful
- Context error if cancelled

### `Release(ctx context.Context) error`

Returns a token to the bucket. Blocks if the bucket is at maximum capacity.

**Parameters:**
- `ctx`: Context for the operation

**Returns:**
- `nil` if successful
- Context error if cancelled

### `SetCapacity(ctx context.Context, c int) error`

Dynamically adjusts the number of available tokens. The new capacity must be between 0 and the maximum capacity.

**Parameters:**
- `ctx`: Context for the operation
- `c`: New capacity (0 <= c <= maxCapacity)

**Returns:**
- `nil` if successful
- Error if capacity is out of bounds or context is cancelled

## Use Cases

- **API Rate Limiting**: Control the rate of API calls to external services
- **Resource Pool Management**: Manage access to limited resources like database connections
- **Concurrent Task Limiting**: Limit the number of concurrent goroutines processing tasks
- **Traffic Shaping**: Control the flow of requests in a distributed system
- **Load Balancing**: Distribute work across workers with dynamic capacity adjustment

## License

MIT License