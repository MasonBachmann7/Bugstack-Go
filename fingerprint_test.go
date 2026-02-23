package bugstack

import (
	"sync"
	"testing"
	"time"
)

func TestGenerateFingerprint(t *testing.T) {
	fp := generateFingerprint("error", "app.go", "handler", 42, "something broke")
	if len(fp) != 16 {
		t.Errorf("expected 16 char fingerprint, got %d", len(fp))
	}
}

func TestFingerprintStable(t *testing.T) {
	a := generateFingerprint("error", "app.go", "handler", 42, "something broke")
	b := generateFingerprint("error", "app.go", "handler", 42, "something broke")
	if a != b {
		t.Error("expected same fingerprint for same inputs")
	}
}

func TestFingerprintDifferent(t *testing.T) {
	a := generateFingerprint("error", "app.go", "handler", 42, "something broke")
	b := generateFingerprint("panic", "app.go", "handler", 42, "something broke")
	if a == b {
		t.Error("expected different fingerprints for different inputs")
	}
}

func TestFingerprintDifferentMessages(t *testing.T) {
	a := generateFingerprint("error", "app.go", "handler", 42, "error one")
	b := generateFingerprint("error", "app.go", "handler", 42, "error two")
	if a == b {
		t.Error("expected different fingerprints for different error messages at same location")
	}
}

func TestExtractErrorLocation(t *testing.T) {
	file, fn, line := extractErrorLocation(1)
	if file == "" {
		t.Error("expected non-empty file")
	}
	if fn == "" {
		t.Error("expected non-empty function")
	}
	if line == 0 {
		t.Error("expected non-zero line")
	}
}

func TestCaptureStack(t *testing.T) {
	stack := captureStack()
	if stack == "" {
		t.Error("expected non-empty stack trace")
	}
	if len(stack) < 10 {
		t.Error("expected substantial stack trace")
	}
}

func TestDeduplicatorFirstAllowed(t *testing.T) {
	d := newDeduplicator(60)
	if !d.shouldSend("fp1") {
		t.Error("expected first occurrence to be allowed")
	}
}

func TestDeduplicatorDuplicate(t *testing.T) {
	d := newDeduplicator(60)
	d.shouldSend("fp1")
	if d.shouldSend("fp1") {
		t.Error("expected duplicate to be blocked")
	}
}

func TestDeduplicatorDifferent(t *testing.T) {
	d := newDeduplicator(60)
	d.shouldSend("fp1")
	if !d.shouldSend("fp2") {
		t.Error("expected different fingerprint to be allowed")
	}
}

func TestDeduplicatorExpired(t *testing.T) {
	d := newDeduplicator(0.01) // 10ms
	d.shouldSend("fp1")
	time.Sleep(20 * time.Millisecond)
	if !d.shouldSend("fp1") {
		t.Error("expected expired entry to be allowed")
	}
}

func TestDeduplicatorClear(t *testing.T) {
	d := newDeduplicator(60)
	d.shouldSend("fp1")
	d.clear()
	if !d.shouldSend("fp1") {
		t.Error("expected cleared entry to be allowed")
	}
}

func TestDeduplicatorConcurrent(t *testing.T) {
	d := newDeduplicator(60)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			d.shouldSend("concurrent")
		}(i)
	}

	wg.Wait()
}
