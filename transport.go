package bugstack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// transport handles sending events to the BugStack API
// via a background goroutine with retry logic.
type transport struct {
	endpoint   string
	apiKey     string
	timeout    time.Duration
	maxRetries int
	debug      bool
	queue      chan map[string]any
	done       chan struct{}
}

func newTransport(endpoint, apiKey string, timeout float64, maxRetries int, debug bool) *transport {
	t := &transport{
		endpoint:   endpoint,
		apiKey:     apiKey,
		timeout:    time.Duration(timeout * float64(time.Second)),
		maxRetries: maxRetries,
		debug:      debug,
		queue:      make(chan map[string]any, 100),
		done:       make(chan struct{}),
	}
	go t.worker()
	return t
}

func (t *transport) worker() {
	for payload := range t.queue {
		t.sendWithRetry(payload)
	}
	close(t.done)
}

func (t *transport) enqueue(payload map[string]any) {
	select {
	case t.queue <- payload:
		// queued
	default:
		if t.debug {
			log.Println("[BugStack] Queue full, dropping event")
		}
	}
}

func (t *transport) sendWithRetry(payload map[string]any) bool {
	body, err := json.Marshal(payload)
	if err != nil {
		if t.debug {
			log.Printf("[BugStack] JSON marshal error: %v", err)
		}
		return false
	}

	client := &http.Client{Timeout: t.timeout}

	for attempt := 0; attempt < t.maxRetries; attempt++ {
		req, err := http.NewRequest("POST", t.endpoint, bytes.NewReader(body))
		if err != nil {
			if t.debug {
				log.Printf("[BugStack] Request creation error: %v", err)
			}
			return false
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-BugStack-API-Key", t.apiKey)
		req.Header.Set("X-BugStack-SDK-Version", Version)

		resp, err := client.Do(req)
		if err != nil {
			if t.debug {
				log.Printf("[BugStack] Send failed (attempt %d): %v", attempt+1, err)
			}
			if attempt < t.maxRetries-1 {
				time.Sleep(time.Duration(1<<uint(attempt)) * time.Second)
			}
			continue
		}

		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode < 400 {
			if t.debug {
				log.Println("[BugStack] Event sent successfully")
			}
			return true
		}

		if t.debug {
			log.Printf("[BugStack] HTTP %d (attempt %d)", resp.StatusCode, attempt+1)
		}

		if attempt < t.maxRetries-1 {
			time.Sleep(time.Duration(1<<uint(attempt)) * time.Second)
		}
	}

	if t.debug {
		log.Println("[BugStack] Max retries exceeded, dropping event")
	}
	return false
}

func (t *transport) shutdown() {
	close(t.queue)
	select {
	case <-t.done:
	case <-time.After(2 * time.Second):
		if t.debug {
			log.Println("[BugStack] Shutdown timeout, some events may be lost")
		}
	}
}

// dryRunSend prints the payload to stderr for debugging.
func dryRunSend(payload map[string]any) {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Printf("[BugStack DryRun] Marshal error: %v\n", err)
		return
	}
	fmt.Printf("[BugStack DryRun] Would send: %s\n", data)
}
