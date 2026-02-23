package bugstack

import (
	"errors"
	"testing"
)

func newTestClient(opts ...func(*Config)) *Client {
	cfg := Config{
		APIKey: "bs_test_key",
		DryRun: true,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	cfg.setDefaults()
	return NewClient(cfg)
}

func TestClientCaptureError(t *testing.T) {
	c := newTestClient()
	defer c.Shutdown()

	err := errors.New("test error")
	result := c.CaptureError(err)
	if !result {
		t.Error("expected capture to succeed")
	}
}

func TestClientCaptureErrorNil(t *testing.T) {
	c := newTestClient()
	defer c.Shutdown()

	result := c.CaptureError(nil)
	if result {
		t.Error("expected false for nil error")
	}
}

func TestClientCaptureMessage(t *testing.T) {
	c := newTestClient()
	defer c.Shutdown()

	result := c.CaptureMessage("something went wrong")
	if !result {
		t.Error("expected capture to succeed")
	}
}

func TestClientCaptureWithRequest(t *testing.T) {
	c := newTestClient()
	defer c.Shutdown()

	err := errors.New("request error")
	result := c.CaptureError(err, WithRequest(&RequestContext{
		Route:  "/api/users",
		Method: "GET",
	}))
	if !result {
		t.Error("expected capture to succeed")
	}
}

func TestClientCaptureWithMetadata(t *testing.T) {
	c := newTestClient()
	defer c.Shutdown()

	err := errors.New("meta error")
	result := c.CaptureError(err, WithMetadata(map[string]any{
		"user_id": "42",
	}))
	if !result {
		t.Error("expected capture to succeed")
	}
}

func TestClientDeduplication(t *testing.T) {
	c := newTestClient()
	defer c.Shutdown()

	err := errors.New("dedup error")
	r1 := c.CaptureError(err)
	r2 := c.CaptureError(err)

	if !r1 {
		t.Error("expected first capture to succeed")
	}
	if r2 {
		t.Error("expected second capture to be deduplicated")
	}
}

func TestClientIgnoredErrors(t *testing.T) {
	c := newTestClient(func(cfg *Config) {
		cfg.IgnoredErrors = []string{"expected"}
	})
	defer c.Shutdown()

	err := errors.New("expected error")
	result := c.CaptureError(err)
	if result {
		t.Error("expected error to be ignored")
	}
}

func TestClientIgnoredNonMatch(t *testing.T) {
	c := newTestClient(func(cfg *Config) {
		cfg.IgnoredErrors = []string{"expected"}
	})
	defer c.Shutdown()

	err := errors.New("unexpected error")
	result := c.CaptureError(err)
	if !result {
		t.Error("expected error to NOT be ignored")
	}
}

func TestClientBeforeSendDrop(t *testing.T) {
	c := newTestClient(func(cfg *Config) {
		cfg.BeforeSend = func(e *Event) *Event {
			return nil
		}
	})
	defer c.Shutdown()

	err := errors.New("dropped error")
	result := c.CaptureError(err)
	if result {
		t.Error("expected event to be dropped by BeforeSend")
	}
}

func TestClientBeforeSendModify(t *testing.T) {
	var captured *Event
	c := newTestClient(func(cfg *Config) {
		cfg.BeforeSend = func(e *Event) *Event {
			e.Metadata = map[string]any{"modified": true}
			captured = e
			return e
		}
	})
	defer c.Shutdown()

	err := errors.New("modify error")
	c.CaptureError(err)

	if captured == nil {
		t.Fatal("expected event to be captured by BeforeSend")
	}
	if captured.Metadata["modified"] != true {
		t.Error("expected metadata to be modified")
	}
}

func TestClientDisabled(t *testing.T) {
	c := newTestClient(func(cfg *Config) {
		cfg.Enabled = Bool(false)
	})
	defer c.Shutdown()

	err := errors.New("disabled error")
	result := c.CaptureError(err)
	if result {
		t.Error("expected capture to fail when disabled")
	}
}

func TestClientShutdownSafe(t *testing.T) {
	c := newTestClient()
	c.Shutdown()
	c.Shutdown() // Double shutdown should not panic
}

func TestClientDryRunNoTransport(t *testing.T) {
	c := newTestClient()
	if c.transport != nil {
		t.Error("expected no transport in dry run mode")
	}
}

func TestClientNeverPanics(t *testing.T) {
	c := newTestClient(func(cfg *Config) {
		cfg.BeforeSend = func(e *Event) *Event {
			panic("hook panic")
		}
	})
	defer c.Shutdown()

	// Should not panic
	err := errors.New("panic test")
	c.CaptureError(err)
}
