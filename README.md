# retry-pkg

A team-focused retry mechanism wrapper for Go applications. Currently provides intelligent HTTP response handling with plans for expanded retry strategies.

## Overview

`retry-pkg` is designed as our team's standardized retry solution, built as a smart wrapper around the proven `avast/retry-go` library. It provides sensible defaults and intelligent error handling while maintaining the flexibility to expand for future use cases.

**Current Focus:** HTTP-aware retry logic with automatic status code handling
**Future Plans:** Expanded to cover database retries, external service calls, and custom retry strategies

## Current Features

- 🌐 **HTTP-Smart Retries**: Automatic handling of HTTP status codes (4xx = no retry, 5xx = retry)
- 🔄 **Sensible Defaults**: 5 retry attempts with proven backoff strategies (this will evolve by using the opts)
- 🎯 **Type Safety**: Generic support for any return type using Go generics
- 🚫 **Unrecoverable Error Detection**: Intelligent distinction between temporary and permanent failures
- 🧪 **Testing Support**: Built-in retry attempt tracking for test verification
- 📦 **Team Standards**: Consistent retry behavior across all team projects

## Installation

```bash
go get TBD
```

## Current Usage

### HTTP Client Retries

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    
    "github.com/KalZyyrk/retry-pkg"
)

func main() {
    ctx := context.Background()
    
    // HTTP GET with automatic retry logic
    resp, err := retries.Retry(ctx, func() (*http.Response, error) {
        return http.Get("https://api.example.com/data")
    })
    
    if err != nil {
        fmt.Printf("Request failed: %v\n", err)
        return
    }
    defer resp.Body.Close()
    
    fmt.Printf("Success! Status: %d\n", resp.StatusCode)
}
```

### Generic Function Retries

```go
// Works with any function signature
result, err := retries.Retry(ctx, func() (string, error) {
    // Your operation here
    return fetchDataFromService()
})
```

## How It Currently Works

### HTTP Response Handling

The package automatically analyzes HTTP responses:

- **2xx-3xx**: Success, no retry needed
- **4xx**: Client errors (Bad Request, Unauthorized, etc.) → **NO RETRY**
  - Marked as unrecoverable since they typically won't succeed on retry
- **5xx**: Server errors → **RETRY ENABLED**
  - Temporary server issues that may resolve with retry

### Default Retry Behavior

- **Attempts**: 5 maximum retries (will evolve)
- **Strategy**: Uses `avast/retry-go` defaults (exponential backoff)
- **Condition**: Only retries recoverable errors
- **Context**: Supports context cancellation

## API Reference

### Core Function

```go
func Retry[T any](ctx context.Context, f func() (T, error), opts ...retry.Option) (T, error)
```

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `f`: Function to execute (returns any type T and error)
- `opts`: Optional `retry-go` options for custom behavior

**Returns:**
- `T`: Result of successful function execution
- `error`: Any error that occurred

### Testing Utilities

```go
func GetCount() int // Returns retry attempts from last Retry() call
```

## Team Usage Examples

### Standard API Call Pattern

```go
func callTeamAPI(endpoint string) (*APIResponse, error) {
    ctx := context.Background()
    
    return retries.Retry(ctx, func() (*APIResponse, error) {
        resp, err := http.Get(fmt.Sprintf("%s%s", baseURL, endpoint))
        if err != nil {
            return nil, err
        }
        defer resp.Body.Close()
        
        var apiResp APIResponse
        if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
            return nil, err
        }
        
        return &apiResp, nil
    })
}
```

### Custom Retry Configuration

```go
import "github.com/avast/retry-go"

// Custom retry behavior for specific needs
result, err := retries.Retry(ctx, yourFunction,
    retry.Attempts(3),                           // Only 3 attempts
    retry.Delay(1*time.Second),                 // 1 second between retries
    retry.DelayType(retry.FixedDelay),          // Fixed delay instead of backoff
)
```

## Testing Your Retries

```go
func TestAPIRetryBehavior(t *testing.T) {
    // Test that 4xx errors don't retry
    _, err := retries.Retry(context.Background(), func() (*http.Response, error) {
        return &http.Response{StatusCode: 400}, nil
    })
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "Bad Request")
    assert.Equal(t, 0, retries.GetCount()) // No retries attempted
    
    // Test that 5xx errors do retry
    callCount := 0
    _, err = retries.Retry(context.Background(), func() (*http.Response, error) {
        callCount++
        if callCount < 3 {
            return &http.Response{StatusCode: 500}, nil
        }
        return &http.Response{StatusCode: 200}, nil
    })
    
    assert.NoError(t, err)
    assert.Equal(t, 2, retries.GetCount()) // 2 retries were made
}
```

## Roadmap & Future Enhancements

### Planned Features
- **GRPC Retry Patterns**: Specialized handling for GRPC connection
- **Circuit Breaker Integration**: Prevent cascading failures in microservice calls  
- **Custom Retry Strategies**: Team-specific retry patterns for different service types
- **Observability**: Metrics and logging integration for retry monitoring
- **Configuration Profiles**: Predefined retry configurations for common team use cases

## Team Guidelines

### When to Use This Package
- ✅ HTTP API calls to external services
- ✅ Any operation that might have transient failures
- ✅ When you need consistent retry behavior across projects

### When NOT to Use
- ❌ Operations that should fail fast (user authentication, validation)
- ❌ Already reliable operations (local file system access)
- ❌ Operations with strict timing requirements

### Best Practices
1. Log retry attempts in production for debugging
2. Use `GetCount()` in tests to verify retry behavior
3. Consider circuit breaker patterns for high-frequency operations

## Contributing (Team Members)

### Adding New Features
1. Create feature branch: `git checkout -b feature/new-retry-pattern`
2. Entrypoint for future use case will still be `checkResAndErr()`
3. Add tests for new functionality
4. Update documentation

## Dependencies

- [`github.com/avast/retry-go`](https://github.com/avast/retry-go) v3.0.0+ - Core retry mechanism
- Go 1.18+ (for generics support)