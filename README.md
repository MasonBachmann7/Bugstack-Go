# bugstack-go

Official Go SDK for [BugStack](https://bugstack.dev) — capture, report, and auto-fix production errors.

[![Go Reference](https://pkg.go.dev/badge/github.com/MasonBachmann7/bugstack-go.svg)](https://pkg.go.dev/github.com/MasonBachmann7/bugstack-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Installation

```bash
go get github.com/MasonBachmann7/bugstack-go
```

## Quick Start

```go
package main

import (
    "errors"
    bugstack "github.com/MasonBachmann7/bugstack-go"
)

func main() {
    bugstack.Init(bugstack.Config{
        APIKey: "bs_live_...",
    })
    defer bugstack.Flush()

    err := riskyOperation()
    if err != nil {
        bugstack.CaptureError(err)
    }
}
```

## Panic Recovery

```go
func handler() {
    defer bugstack.Recover()
    // ... code that might panic
}
```

## HTTP Middleware

### net/http

```go
import "github.com/MasonBachmann7/bugstack-go/middleware"

mux := http.NewServeMux()
mux.HandleFunc("/", handler)
http.ListenAndServe(":8080", middleware.NetHTTP(mux))
```

### Gin

```go
import "github.com/MasonBachmann7/bugstack-go/middleware"

r := gin.Default()
r.Use(gin.HandlerFunc(middleware.GinRecovery()))
```

### Echo

```go
import "github.com/MasonBachmann7/bugstack-go/middleware"

e := echo.New()
e.Use(echo.MiddlewareFunc(middleware.EchoRecover()))
```

## Configuration

```go
bugstack.Init(bugstack.Config{
    APIKey:              "bs_live_...",       // Required
    Environment:         "production",       // Default: "production"
    AutoFix:             true,               // Enable AI-powered auto-fix
    Debug:               false,              // Log SDK activity
    DryRun:              false,              // Log without sending
    Enabled:             true,               // Kill switch
    DeduplicationWindow: 300,                // Seconds (default: 5 min)
    Timeout:             5.0,                // HTTP timeout in seconds
    MaxRetries:          3,                  // Retry attempts
    IgnoredErrors:       []string{           // Error substrings to skip
        "context canceled",
    },
    BeforeSend: func(e *bugstack.Event) *bugstack.Event {
        // Return nil to drop, modify and return to keep
        return e
    },
})
```

## Data Transparency

### `BeforeSend` Hook

Inspect, modify, or drop any event before it leaves your application:

```go
bugstack.Init(bugstack.Config{
    APIKey: "bs_live_...",
    BeforeSend: func(e *bugstack.Event) *bugstack.Event {
        if e.Request != nil && e.Request.Route == "/health" {
            return nil // Drop health check errors
        }
        delete(e.Metadata, "secret")
        return e
    },
})
```

### `DryRun` Mode

See exactly what would be sent without network requests:

```go
bugstack.Init(bugstack.Config{
    APIKey: "bs_live_...",
    DryRun: true,
})
// Prints: [BugStack DryRun] Would send: { ... }
```

## What Gets Sent

```json
{
  "apiKey": "bs_live_...",
  "error": {
    "message": "runtime error: index out of range",
    "stackTrace": "goroutine 1 [running]: ...",
    "file": "main.go",
    "function": "main.handler",
    "fingerprint": "a1b2c3d4e5f6g7h8"
  },
  "environment": {
    "language": "go",
    "languageVersion": "go1.22.0",
    "os": "linux",
    "sdkVersion": "1.0.0"
  },
  "timestamp": "2026-01-15T08:30:00Z"
}
```

Zero external dependencies. No cookies, IP addresses, or user data collected.

## License

MIT — see [LICENSE](LICENSE).
