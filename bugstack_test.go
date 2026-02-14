package bugstack

import (
	"testing"
)

func resetGlobal() {
	mu.Lock()
	defer mu.Unlock()
	if globalClient != nil {
		globalClient.Shutdown()
	}
	globalClient = nil
}

func TestInit(t *testing.T) {
	defer resetGlobal()

	Init(Config{
		APIKey: "bs_test_key",
		DryRun: true,
	})

	if GetClient() == nil {
		t.Fatal("expected client to be initialized")
	}
}

func TestInitSetsDefaults(t *testing.T) {
	defer resetGlobal()

	Init(Config{
		APIKey: "bs_test_key",
		DryRun: true,
	})

	c := GetClient()
	if c.config.Endpoint != "https://api.bugstack.dev/api/capture" {
		t.Errorf("expected default endpoint, got %s", c.config.Endpoint)
	}
	if c.config.Environment != "production" {
		t.Errorf("expected default environment, got %s", c.config.Environment)
	}
	if c.config.DeduplicationWindow != 300 {
		t.Errorf("expected default dedup window 300, got %f", c.config.DeduplicationWindow)
	}
	if c.config.Timeout != 5 {
		t.Errorf("expected default timeout 5, got %f", c.config.Timeout)
	}
	if c.config.MaxRetries != 3 {
		t.Errorf("expected default max retries 3, got %d", c.config.MaxRetries)
	}
}

func TestCaptureErrorWithoutInit(t *testing.T) {
	resetGlobal()
	result := CaptureError(nil)
	if result {
		t.Error("expected false when client not initialized")
	}
}

func TestCaptureErrorNil(t *testing.T) {
	defer resetGlobal()

	Init(Config{
		APIKey: "bs_test_key",
		DryRun: true,
	})

	result := CaptureError(nil)
	if result {
		t.Error("expected false for nil error")
	}
}

func TestCaptureMessageWithoutInit(t *testing.T) {
	resetGlobal()
	result := CaptureMessage("test")
	if result {
		t.Error("expected false when client not initialized")
	}
}

func TestGetClientNil(t *testing.T) {
	resetGlobal()
	if GetClient() != nil {
		t.Error("expected nil before init")
	}
}

func TestFlushSafe(t *testing.T) {
	// Should not panic
	resetGlobal()
	Flush()
}

func TestReinitShutsPrevious(t *testing.T) {
	defer resetGlobal()

	Init(Config{
		APIKey: "bs_test_key_1",
		DryRun: true,
	})
	c1 := GetClient()

	Init(Config{
		APIKey: "bs_test_key_2",
		DryRun: true,
	})
	c2 := GetClient()

	if c1 == c2 {
		t.Error("expected different client instances after reinit")
	}
}
