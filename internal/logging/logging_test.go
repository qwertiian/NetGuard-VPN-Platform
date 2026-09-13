package logging

import (
	"testing"
)

func TestInit(t *testing.T) {
	Init("DEBUG")
	if Logger == nil {
		t.Error("Logger should not be nil after Init")
	}

	Init("INVALID")
	if Logger == nil {
		t.Error("Logger should fallback to INFO and not be nil")
	}
}

func TestScrubSecrets(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no secrets",
			input:    "just a normal string",
			expected: "just a normal string",
		},
		{
			name:     "contains wireguard key",
			input:    "private key: cI2v/abcdefghijklmnopqrstuvwxyz0123456789+A=",
			expected: "private key: [REDACTED]",
		},
		{
			name:     "multiple keys",
			input:    "key1: cI2v/abcdefghijklmnopqrstuvwxyz0123456789+A= key2: cI2v/abcdefghijklmnopqrstuvwxyz0123456789+A=",
			expected: "key1: [REDACTED] key2: [REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ScrubSecrets(tt.input)
			if result != tt.expected {
				t.Errorf("ScrubSecrets() = %v, want %v", result, tt.expected)
			}
		})
	}
}
