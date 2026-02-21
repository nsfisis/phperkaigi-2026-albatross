package tournament

import "testing"

func TestStandardBracketSeeds(t *testing.T) {
	tests := []struct {
		name        string
		bracketSize int
		expected    []int
	}{
		{
			name:        "bracket_size=2",
			bracketSize: 2,
			expected:    []int{1, 2},
		},
		{
			name:        "bracket_size=4",
			bracketSize: 4,
			expected:    []int{1, 4, 2, 3},
		},
		{
			name:        "bracket_size=8",
			bracketSize: 8,
			expected:    []int{1, 8, 4, 5, 2, 7, 3, 6},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StandardBracketSeeds(tt.bracketSize)
			if len(got) != len(tt.expected) {
				t.Fatalf("expected length %d, got %d", len(tt.expected), len(got))
			}
			for i, v := range tt.expected {
				if got[i] != v {
					t.Errorf("position %d: expected seed %d, got %d", i, v, got[i])
				}
			}
		})
	}
}

func TestStandardBracketSeeds_Seed1And2OppositeSides(t *testing.T) {
	seeds := StandardBracketSeeds(8)
	seed1Pos := -1
	seed2Pos := -1
	for i, s := range seeds {
		if s == 1 {
			seed1Pos = i
		}
		if s == 2 {
			seed2Pos = i
		}
	}
	if seed1Pos >= 4 {
		t.Errorf("Seed 1 should be in first half, but at position %d", seed1Pos)
	}
	if seed2Pos < 4 {
		t.Errorf("Seed 2 should be in second half, but at position %d", seed2Pos)
	}
}

func TestStandardBracketSeeds_AllSeedsPresent(t *testing.T) {
	for _, size := range []int{2, 4, 8, 16} {
		seeds := StandardBracketSeeds(size)
		seen := make(map[int]bool)
		for _, s := range seeds {
			if s < 1 || s > size {
				t.Errorf("bracket_size=%d: seed %d out of range", size, s)
			}
			if seen[s] {
				t.Errorf("bracket_size=%d: duplicate seed %d", size, s)
			}
			seen[s] = true
		}
		if len(seen) != size {
			t.Errorf("bracket_size=%d: expected %d unique seeds, got %d", size, size, len(seen))
		}
	}
}

func TestFindSeedByUserID(t *testing.T) {
	entries := []Entry{
		{User: Player{UserID: 10}, Seed: 1},
		{User: Player{UserID: 20}, Seed: 2},
		{User: Player{UserID: 30}, Seed: 3},
	}

	if got := findSeedByUserID(entries, 10); got != 1 {
		t.Errorf("expected seed 1 for user 10, got %d", got)
	}
	if got := findSeedByUserID(entries, 20); got != 2 {
		t.Errorf("expected seed 2 for user 20, got %d", got)
	}
	if got := findSeedByUserID(entries, 999); got != 0 {
		t.Errorf("expected seed 0 for unknown user, got %d", got)
	}
}

func TestNextPowerOf2(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{2, 2},
		{3, 4},
		{4, 4},
		{5, 8},
		{6, 8},
		{7, 8},
		{8, 8},
		{9, 16},
	}
	for _, tt := range tests {
		got := nextPowerOf2(tt.input)
		if got != tt.expected {
			t.Errorf("nextPowerOf2(%d) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestLog2Int(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{1, 0},
		{2, 1},
		{4, 2},
		{8, 3},
		{16, 4},
	}
	for _, tt := range tests {
		got := log2Int(tt.input)
		if got != tt.expected {
			t.Errorf("log2Int(%d) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}
