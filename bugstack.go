// Package bugstack provides the official Go SDK for BugStack —
// capture, report, and auto-fix production errors.
//
// Usage:
//
//	import "github.com/MasonBachmann7/bugstack-go"
//
//	func main() {
//	    bugstack.Init(bugstack.Config{
//	        APIKey: "bs_live_...",
//	    })
//	    defer bugstack.Flush()
//
//	    // errors are captured automatically via middleware,
//	    // or manually:
//	    err := riskyOperation()
//	    if err != nil {
//	        bugstack.CaptureError(err)
//	    }
//	}
package bugstack

import (
	"sync"
)

// Version is the SDK version.
const Version = "1.1.0"

var (
	globalClient *Client
	mu           sync.RWMutex
)

// Init initializes the global BugStack client.
// Call this once at application startup.
func Init(cfg Config) {
	cfg.setDefaults()

	mu.Lock()
	defer mu.Unlock()

	if globalClient != nil {
		globalClient.Shutdown()
	}
	globalClient = NewClient(cfg)
}

// CaptureError captures an error and sends it to BugStack.
// Returns true if the error was accepted (not filtered/deduplicated).
func CaptureError(err error, opts ...CaptureOption) bool {
	mu.RLock()
	c := globalClient
	mu.RUnlock()

	if c == nil {
		return false
	}
	return c.CaptureError(err, opts...)
}

// CaptureMessage captures a message string as an event.
func CaptureMessage(msg string, opts ...CaptureOption) bool {
	mu.RLock()
	c := globalClient
	mu.RUnlock()

	if c == nil {
		return false
	}
	return c.CaptureMessage(msg, opts...)
}

// Recover captures a panic value. Call this in a deferred function.
//
//	defer bugstack.Recover()
func Recover(opts ...CaptureOption) {
	if r := recover(); r != nil {
		mu.RLock()
		c := globalClient
		mu.RUnlock()

		if c != nil {
			c.RecoverPanic(r, opts...)
		}
	}
}

// GetClient returns the global client instance, or nil if not initialized.
func GetClient() *Client {
	mu.RLock()
	defer mu.RUnlock()
	return globalClient
}

// Flush flushes pending events and shuts down the transport.
// Call this before application exit.
func Flush() {
	mu.Lock()
	defer mu.Unlock()

	if globalClient != nil {
		globalClient.Shutdown()
	}
}
