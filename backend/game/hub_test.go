package game

import "testing"

func TestCalcCodeSize_PHP(t *testing.T) {
	hub := &Hub{}
	tests := []struct {
		name     string
		code     string
		language string
		want     int
	}{
		{
			name:     "simple php code",
			code:     "<?php echo 1;",
			language: "php",
			want:     6, // "echo1;" after stripping whitespace and "<?php"
		},
		{
			name:     "php with short open tag",
			code:     "<? echo 1;",
			language: "php",
			want:     6, // "echo1;" after stripping whitespace and "<?"
		},
		{
			name:     "php with closing tag",
			code:     "<?php echo 1; ?>",
			language: "php",
			want:     6, // "echo1;" after stripping whitespace, "<?php", and "?>"
		},
		{
			name:     "php with whitespace",
			code:     "<?php echo   1 ;  ?>",
			language: "php",
			want:     6,
		},
		{
			name:     "non-php language",
			code:     "print(1)",
			language: "swift",
			want:     8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hub.CalcCodeSize(tt.code, tt.language)
			if got != tt.want {
				t.Errorf("CalcCodeSize(%q, %q) = %d, want %d", tt.code, tt.language, got, tt.want)
			}
		})
	}
}

func TestIsTestcaseResultCorrect(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		actual   string
		want     bool
	}{
		{
			name:     "exact match",
			expected: "hello",
			actual:   "hello",
			want:     true,
		},
		{
			name:     "trailing newline ignored",
			expected: "hello\n",
			actual:   "hello",
			want:     true,
		},
		{
			name:     "CRLF normalized",
			expected: "hello\r\n",
			actual:   "hello\n",
			want:     true,
		},
		{
			name:     "mismatch",
			expected: "hello",
			actual:   "world",
			want:     false,
		},
		{
			name:     "multiline match",
			expected: "line1\nline2",
			actual:   "line1\nline2\n",
			want:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTestcaseResultCorrect(tt.expected, tt.actual)
			if got != tt.want {
				t.Errorf("isTestcaseResultCorrect(%q, %q) = %v, want %v", tt.expected, tt.actual, got, tt.want)
			}
		})
	}
}
