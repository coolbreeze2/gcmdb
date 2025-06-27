package deployment

import (
	"testing"
)

func TestTruncNameLeft63(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Short name, no digits",
			input:    "my-app-name",
			expected: "my-app-name",
		},
		{
			name:     "Name with digit prefix and dash",
			input:    "123-abc-app",
			expected: "abc-app",
		},
		{
			name:     "Name longer than 63 chars",
			input:    "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-0123456789",
			expected: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-0123456789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncNameLeft63(tt.input)
			if got != tt.expected {
				t.Errorf("truncNameLeft63(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
