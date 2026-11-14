package handlers

import (
	"testing"
	"time"
)

func TestGenerateResetCode(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		code, err := generateResetCode()
		if err != nil {
			t.Fatalf("generateResetCode error: %v", err)
		}
		if len(code) != resetCodeLength {
			t.Fatalf("code %q has length %d, want %d", code, len(code), resetCodeLength)
		}
		if seen[code] {
			t.Fatalf("duplicate code generated: %q", code)
		}
		seen[code] = true
	}
}

func TestResetCodeStoreConsumeOnce(t *testing.T) {
	storeResetCode("ABC12345", "a@b.com")
	email, ok := consumeResetCode("ABC12345")
	if !ok || email != "a@b.com" {
		t.Fatalf("first consume failed: ok=%v email=%q", ok, email)
	}
	if _, ok := consumeResetCode("ABC12345"); ok {
		t.Error("code was accepted twice")
	}
	if _, ok := consumeResetCode("UNKNOWN0"); ok {
		t.Error("unknown code was accepted")
	}
}

func TestResetCodeStoreExpiry(t *testing.T) {
	passwordResetStore.Lock()
	passwordResetStore.codes["EXPIRED1"] = resetEntry{email: "x@y.com", expiresAt: time.Now().Add(-time.Second)}
	passwordResetStore.Unlock()

	if _, ok := consumeResetCode("EXPIRED1"); ok {
		t.Error("expired code was accepted")
	}
}
