package bugstack

import (
	"testing"
	"time"
)

func TestEventToPayload(t *testing.T) {
	cfg := Config{
		APIKey:      "bs_test_key",
		Environment: "production",
	}
	cfg.setDefaults()

	event := &Event{
		Message:       "test error",
		StackTrace:    "goroutine 1 ...",
		File:          "main.go",
		Function:      "main.handler",
		Fingerprint:   "abc123",
		ExceptionType: "error",
		Environment: EnvironmentInfo{
			Language:        "go",
			LanguageVersion: "go1.21",
			OS:              "linux",
			SDKVersion:      Version,
		},
		Timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	payload := event.toPayload(cfg)

	if payload["apiKey"] != "bs_test_key" {
		t.Errorf("expected apiKey bs_test_key, got %v", payload["apiKey"])
	}

	errMap := payload["error"].(map[string]any)
	if errMap["message"] != "test error" {
		t.Errorf("expected message 'test error', got %v", errMap["message"])
	}
	if errMap["fingerprint"] != "abc123" {
		t.Errorf("expected fingerprint abc123, got %v", errMap["fingerprint"])
	}

	envMap := payload["environment"].(map[string]any)
	if envMap["language"] != "go" {
		t.Errorf("expected language go, got %v", envMap["language"])
	}
	if envMap["sdkVersion"] != Version {
		t.Errorf("expected sdkVersion %s, got %v", Version, envMap["sdkVersion"])
	}
}

func TestEventToPayloadWithRequest(t *testing.T) {
	cfg := Config{APIKey: "bs_test_key"}
	cfg.setDefaults()

	event := &Event{
		Message: "err",
		Request: &RequestContext{
			Route:  "/api/users",
			Method: "GET",
		},
		Timestamp: time.Now(),
	}

	payload := event.toPayload(cfg)

	reqMap, ok := payload["request"].(map[string]any)
	if !ok {
		t.Fatal("expected request in payload")
	}
	if reqMap["route"] != "/api/users" {
		t.Errorf("expected route /api/users, got %v", reqMap["route"])
	}
	if reqMap["method"] != "GET" {
		t.Errorf("expected method GET, got %v", reqMap["method"])
	}
}

func TestEventToPayloadWithProjectID(t *testing.T) {
	cfg := Config{APIKey: "bs_test_key", ProjectID: "proj_123"}
	cfg.setDefaults()

	event := &Event{Message: "err", Timestamp: time.Now()}
	payload := event.toPayload(cfg)

	if payload["projectId"] != "proj_123" {
		t.Errorf("expected projectId proj_123, got %v", payload["projectId"])
	}
}

func TestEventToPayloadAutoFix(t *testing.T) {
	cfg := Config{APIKey: "bs_test_key", AutoFix: true}
	cfg.setDefaults()

	event := &Event{Message: "err", Timestamp: time.Now()}
	payload := event.toPayload(cfg)

	meta, ok := payload["metadata"].(map[string]any)
	if !ok {
		t.Fatal("expected metadata in payload")
	}
	if meta["autoFix"] != true {
		t.Error("expected autoFix=true in metadata")
	}
}

func TestEventToPayloadWithMetadata(t *testing.T) {
	cfg := Config{APIKey: "bs_test_key"}
	cfg.setDefaults()

	event := &Event{
		Message:   "err",
		Metadata:  map[string]any{"user": "42"},
		Timestamp: time.Now(),
	}
	payload := event.toPayload(cfg)

	meta, ok := payload["metadata"].(map[string]any)
	if !ok {
		t.Fatal("expected metadata in payload")
	}
	if meta["user"] != "42" {
		t.Errorf("expected user=42, got %v", meta["user"])
	}
}
