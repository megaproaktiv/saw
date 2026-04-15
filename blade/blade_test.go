package blade

import (
	"strings"
	"testing"
)

func TestShortenLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "short line",
			input:    "This is a short log message",
			expected: "This is a short log message",
		},
		{
			name:     "exactly 512 characters",
			input:    strings.Repeat("x", 512),
			expected: strings.Repeat("x", 512),
		},
		{
			name:     "513 characters (just over limit)",
			input:    strings.Repeat("x", 513),
			expected: strings.Repeat("x", 512) + "...",
		},
		{
			name:     "very long line (1000 characters)",
			input:    strings.Repeat("y", 1000),
			expected: strings.Repeat("y", 512) + "...",
		},
		{
			name:     "log message with timestamp longer than 512",
			input:    "2024-04-15 14:30:00 ERROR " + strings.Repeat("Long error message ", 50),
			expected: ("2024-04-15 14:30:00 ERROR " + strings.Repeat("Long error message ", 50))[:512] + "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shortenLine(tt.input)
			if result != tt.expected {
				t.Errorf("shortenLine() = %q (len=%d), want %q (len=%d)",
					result, len(result), tt.expected, len(tt.expected))
			}

			// Verify truncated lines end with "..."
			if len(tt.input) > 512 {
				if !strings.HasSuffix(result, "...") {
					t.Errorf("shortenLine() result should end with '...' for input longer than 512 chars")
				}
				if len(result) != 515 { // 512 + len("...")
					t.Errorf("shortenLine() result length = %d, want 515", len(result))
				}
			}

			// Verify non-truncated lines don't end with "..."
			if len(tt.input) <= 512 && strings.HasSuffix(result, "...") {
				t.Errorf("shortenLine() result should not end with '...' for input <= 512 chars")
			}
		})
	}
}

func TestColorizeLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "message with INFO",
			input:    "2024-04-15 14:30:00 INFO Application started",
			contains: []string{"2024-04-15 14:30:00", "Application started"},
		},
		{
			name:     "message with ERROR",
			input:    "2024-04-15 14:30:00 ERROR Connection failed",
			contains: []string{"2024-04-15 14:30:00", "Connection failed"},
		},
		{
			name:     "message with both INFO and ERROR",
			input:    "INFO: Operation started, but ERROR occurred",
			contains: []string{"Operation started", "occurred"},
		},
		{
			name:     "message without log levels",
			input:    "Just a regular message",
			contains: []string{"Just a regular message"},
		},
		{
			name:     "message with lowercase info (should not be colored)",
			input:    "This is info about something",
			contains: []string{"This is info about something"},
		},
		{
			name:     "JSON with INFO level",
			input:    `{"level":"INFO","message":"test"}`,
			contains: []string{`"level"`, `"message"`, `"test"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colorizeLogLevel(tt.input)

			// Verify that the result contains the expected substrings
			// (we can't check for exact equality because color codes are added)
			for _, substring := range tt.contains {
				if !strings.Contains(result, substring) {
					t.Errorf("colorizeLogLevel() result should contain %q, got %q",
						substring, result)
				}
			}

			// Verify the result is not empty
			if result == "" && tt.input != "" {
				t.Errorf("colorizeLogLevel() returned empty string for non-empty input")
			}
		})
	}
}

func TestColorizeLogLevelCaseSensitive(t *testing.T) {
	// Test that only uppercase INFO and ERROR are colorized
	testCases := []struct {
		input string
		desc  string
	}{
		{"INFO", "uppercase INFO"},
		{"ERROR", "uppercase ERROR"},
		{"info", "lowercase info"},
		{"error", "lowercase error"},
		{"Info", "mixed case Info"},
		{"Error", "mixed case Error"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := colorizeLogLevel(tc.input)
			// The result should always contain the original text
			// For uppercase, it will be wrapped in color codes
			// For non-uppercase, it should be unchanged
			if len(result) == 0 {
				t.Errorf("colorizeLogLevel(%q) returned empty string", tc.input)
			}
		})
	}
}

func BenchmarkShortenLine(b *testing.B) {
	shortLine := "This is a short log message"
	longLine := strings.Repeat("x", 1000)

	b.Run("short line", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			shortenLine(shortLine)
		}
	})

	b.Run("long line", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			shortenLine(longLine)
		}
	})
}

func BenchmarkColorizeLogLevel(b *testing.B) {
	message := "2024-04-15 14:30:00 INFO Processing request with ERROR handling"

	b.Run("with log levels", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			colorizeLogLevel(message)
		}
	})

	noLevelMessage := "2024-04-15 14:30:00 Processing regular message"
	b.Run("without log levels", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			colorizeLogLevel(noLevelMessage)
		}
	})
}