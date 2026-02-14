package bugstack

import (
	"crypto/sha256"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

// generateFingerprint creates a stable SHA-256 fingerprint for an error.
func generateFingerprint(errType, file, function string, line int) string {
	key := fmt.Sprintf("%s:%s:%s:%d", errType, file, function, line)
	h := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", h[:8]) // 16 hex chars
}

// extractErrorLocation extracts file, function, and line from the call stack.
// skip controls how many stack frames to skip.
func extractErrorLocation(skip int) (file string, function string, line int) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", "", 0
	}

	fn := runtime.FuncForPC(pc)
	if fn != nil {
		function = fn.Name()
		// Shorten "github.com/user/pkg.Func" to "pkg.Func"
		if idx := strings.LastIndex(function, "/"); idx >= 0 {
			function = function[idx+1:]
		}
	}

	return file, function, line
}

// captureStack captures the full goroutine stack trace.
func captureStack() string {
	buf := make([]byte, 4096)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return string(buf[:n])
		}
		buf = make([]byte, len(buf)*2)
	}
}

// deduplicator prevents the same error from being reported
// multiple times within a configurable time window.
type deduplicator struct {
	cache  map[string]time.Time
	window time.Duration
	mu     sync.Mutex
}

func newDeduplicator(windowSeconds float64) *deduplicator {
	return &deduplicator{
		cache:  make(map[string]time.Time),
		window: time.Duration(windowSeconds * float64(time.Second)),
	}
}

// shouldSend returns true if the fingerprint hasn't been seen recently.
func (d *deduplicator) shouldSend(fingerprint string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()

	if lastSent, ok := d.cache[fingerprint]; ok {
		if now.Sub(lastSent) < d.window {
			return false
		}
	}

	d.cache[fingerprint] = now
	d.cleanup(now)
	return true
}

func (d *deduplicator) cleanup(now time.Time) {
	for fp, ts := range d.cache {
		if now.Sub(ts) >= d.window {
			delete(d.cache, fp)
		}
	}
}

func (d *deduplicator) clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cache = make(map[string]time.Time)
}
