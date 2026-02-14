package bugstack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestTransportSendsPayload(t *testing.T) {
	var received atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received.Store(true)
		w.WriteHeader(200)
	}))
	defer server.Close()

	tr := newTransport(server.URL, "bs_test_key", 5, 1, false)
	tr.enqueue(map[string]any{"test": true})

	time.Sleep(1 * time.Second)
	tr.shutdown()

	if !received.Load() {
		t.Error("expected payload to be received by server")
	}
}

func TestTransportSendsCorrectHeaders(t *testing.T) {
	var apiKey, sdkVersion, contentType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey = r.Header.Get("X-BugStack-API-Key")
		sdkVersion = r.Header.Get("X-BugStack-SDK-Version")
		contentType = r.Header.Get("Content-Type")
		w.WriteHeader(200)
	}))
	defer server.Close()

	tr := newTransport(server.URL, "bs_test_key", 5, 1, false)
	tr.enqueue(map[string]any{"test": true})

	time.Sleep(1 * time.Second)
	tr.shutdown()

	if apiKey != "bs_test_key" {
		t.Errorf("expected API key header, got %s", apiKey)
	}
	if sdkVersion != Version {
		t.Errorf("expected SDK version header, got %s", sdkVersion)
	}
	if contentType != "application/json" {
		t.Errorf("expected JSON content type, got %s", contentType)
	}
}

func TestTransportSendsValidJSON(t *testing.T) {
	var receivedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 4096)
		n, _ := r.Body.Read(buf)
		receivedBody = buf[:n]
		w.WriteHeader(200)
	}))
	defer server.Close()

	tr := newTransport(server.URL, "bs_test_key", 5, 1, false)
	tr.enqueue(map[string]any{"apiKey": "test", "error": map[string]any{"message": "hello"}})

	time.Sleep(1 * time.Second)
	tr.shutdown()

	var parsed map[string]any
	err := json.Unmarshal(receivedBody, &parsed)
	if err != nil {
		t.Errorf("expected valid JSON body, got error: %v", err)
	}
}

func TestTransportRetry(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := callCount.Add(1)
		if count <= 1 {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	tr := newTransport(server.URL, "bs_test_key", 5, 3, false)
	tr.enqueue(map[string]any{"test": true})

	time.Sleep(5 * time.Second)
	tr.shutdown()

	if callCount.Load() < 2 {
		t.Errorf("expected at least 2 attempts, got %d", callCount.Load())
	}
}

func TestTransportQueueBounded(t *testing.T) {
	tr := newTransport("http://localhost:0", "bs_test_key", 5, 1, false)

	// Enqueue more than buffer size
	for i := 0; i < 150; i++ {
		tr.enqueue(map[string]any{"index": i})
	}

	// Queue should be bounded at 100
	if len(tr.queue) > 100 {
		t.Errorf("expected queue bounded at 100, got %d", len(tr.queue))
	}

	tr.shutdown()
}

func TestTransportShutdown(t *testing.T) {
	tr := newTransport("http://localhost:0", "bs_test_key", 5, 1, false)
	tr.shutdown()
	// Should not block or panic
}
