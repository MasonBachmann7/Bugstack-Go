package bugstack

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"
)

// Client is the core BugStack client. It handles capturing errors,
// deduplication, filtering, and transport.
type Client struct {
	config       Config
	transport    *transport
	deduplicator *deduplicator
	enabled      bool
}

// NewClient creates a new BugStack client with the given configuration.
func NewClient(cfg Config) *Client {
	cfg.setDefaults()

	c := &Client{
		config:       cfg,
		deduplicator: newDeduplicator(cfg.DeduplicationWindow),
		enabled:      *cfg.Enabled, // setDefaults ensures Enabled is non-nil (defaults to true)
	}

	if c.enabled && !cfg.DryRun {
		c.transport = newTransport(
			cfg.Endpoint,
			cfg.APIKey,
			cfg.Timeout,
			cfg.MaxRetries,
			cfg.Debug,
		)
	}

	if cfg.Debug {
		log.Printf("[BugStack] Client initialized (endpoint=%s, dry_run=%v)", cfg.Endpoint, cfg.DryRun)
	}

	return c
}

// CaptureError captures an error and sends it to BugStack.
// Returns true if the error was accepted.
func (c *Client) CaptureError(err error, opts ...CaptureOption) bool {
	if err == nil {
		return false
	}

	defer func() {
		if r := recover(); r != nil && c.config.Debug {
			log.Printf("[BugStack] Panic during capture: %v", r)
		}
	}()

	return c.doCapture(err.Error(), "error", 3, opts...)
}

// CaptureMessage captures a message string as an event.
func (c *Client) CaptureMessage(msg string, opts ...CaptureOption) bool {
	defer func() {
		if r := recover(); r != nil && c.config.Debug {
			log.Printf("[BugStack] Panic during capture: %v", r)
		}
	}()

	return c.doCapture(msg, "message", 3, opts...)
}

// RecoverPanic captures a panic value. Used internally by Recover().
func (c *Client) RecoverPanic(r any, opts ...CaptureOption) {
	msg := fmt.Sprintf("panic: %v", r)
	c.doCapture(msg, "panic", 4, opts...)
}

func (c *Client) doCapture(message, errType string, skip int, opts ...CaptureOption) bool {
	if !c.enabled {
		return false
	}

	// Check ignored errors
	if c.isIgnored(message) {
		if c.config.Debug {
			log.Printf("[BugStack] Error ignored: %s", message)
		}
		return false
	}

	// Apply options
	o := &captureOptions{}
	for _, opt := range opts {
		opt(o)
	}

	// Extract location
	file, function, line := extractErrorLocation(skip)
	stack := captureStack()

	// Build fingerprint (includes message so wrapper call sites produce distinct fingerprints)
	fp := generateFingerprint(errType, file, function, line, message)

	// Build event
	event := &Event{
		Message:       message,
		StackTrace:    stack,
		File:          file,
		Function:      function,
		Fingerprint:   fp,
		ExceptionType: errType,
		Request:       o.request,
		Environment: EnvironmentInfo{
			Language:        "go",
			LanguageVersion: runtime.Version(),
			OS:              runtime.GOOS,
			SDKVersion:      Version,
		},
		Timestamp: time.Now().UTC(),
		Metadata:  o.metadata,
	}

	// before_send hook
	if c.config.BeforeSend != nil {
		event = c.config.BeforeSend(event)
		if event == nil {
			if c.config.Debug {
				log.Println("[BugStack] Event dropped by BeforeSend")
			}
			return false
		}
	}

	// Deduplication
	if !c.deduplicator.shouldSend(event.Fingerprint) {
		if c.config.Debug {
			log.Printf("[BugStack] Event deduplicated: %s", event.Fingerprint)
		}
		return false
	}

	// Build payload
	payload := event.toPayload(c.config)

	// Dry run
	if c.config.DryRun {
		dryRunSend(payload)
		return true
	}

	// Enqueue
	if c.transport != nil {
		c.transport.enqueue(payload)
	}

	if c.config.Debug {
		log.Printf("[BugStack] Event queued: %s", event.Fingerprint)
	}

	return true
}

func (c *Client) isIgnored(message string) bool {
	for _, pattern := range c.config.IgnoredErrors {
		if strings.Contains(message, pattern) {
			return true
		}
	}
	return false
}

// Shutdown flushes pending events and stops the transport.
func (c *Client) Shutdown() {
	if c.transport != nil {
		c.transport.shutdown()
		c.transport = nil
	}
}
