package auth

import (
	"testing"
)

func TestGenerateSessionID(t *testing.T) {
	id, err := GenerateSessionID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 32 bytes → 64 hex characters
	if len(id) != 64 {
		t.Errorf("expected session ID length 64, got %d", len(id))
	}
}

func TestGenerateSessionID_Unique(t *testing.T) {
	id1, err := GenerateSessionID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	id2, err := GenerateSessionID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id1 == id2 {
		t.Error("expected unique session IDs, got identical values")
	}
}

func TestHashSessionID(t *testing.T) {
	raw := "abc123"
	hashed := HashSessionID(raw)
	// SHA-256 produces 32 bytes → 64 hex characters
	if len(hashed) != 64 {
		t.Errorf("expected hash length 64, got %d", len(hashed))
	}
	// Same input should produce same hash
	if hashed != HashSessionID(raw) {
		t.Error("expected deterministic hash")
	}
	// Different input should produce different hash
	if hashed == HashSessionID("different") {
		t.Error("expected different hashes for different inputs")
	}
}
