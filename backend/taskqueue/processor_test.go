package taskqueue

import (
	"testing"
)

func TestCalcCodeHash(t *testing.T) {
	// Same code + same testcaseID should produce same hash
	hash1 := calcCodeHash("echo hello", 1)
	hash2 := calcCodeHash("echo hello", 1)
	if hash1 != hash2 {
		t.Errorf("same input produced different hashes: %q vs %q", hash1, hash2)
	}

	// Different code should produce different hash
	hash3 := calcCodeHash("echo world", 1)
	if hash1 == hash3 {
		t.Errorf("different code produced same hash: %q", hash1)
	}

	// Different testcaseID should produce different hash
	hash4 := calcCodeHash("echo hello", 2)
	if hash1 == hash4 {
		t.Errorf("different testcaseID produced same hash: %q", hash1)
	}

	// Hash should be a valid hex md5 (32 characters)
	if len(hash1) != 32 {
		t.Errorf("hash length = %d, want 32", len(hash1))
	}
}

func TestCalcCodeHash_EmptyCode(t *testing.T) {
	hash := calcCodeHash("", 0)
	if hash == "" {
		t.Error("hash should not be empty for empty code")
	}
	if len(hash) != 32 {
		t.Errorf("hash length = %d, want 32", len(hash))
	}
}
