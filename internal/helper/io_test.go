package helper

import (
	"testing"
)

func TestFormatStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		expected string
	}{
		{
			name:     "nil slice",
			slice:    nil,
			expected: "null",
		},
		{
			name:     "empty slice",
			slice:    []string{},
			expected: "null",
		},
		{
			name:     "single item",
			slice:    []string{"https://example.com/crl"},
			expected: "https://example.com/crl",
		},
		{
			name:     "multiple items",
			slice:    []string{"https://example.com/crl1", "https://example.com/crl2"},
			expected: "https://example.com/crl1, https://example.com/crl2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatStringSlice(tt.slice)
			if result != tt.expected {
				t.Errorf("formatStringSlice() = %v, want %v", result, tt.expected)
			}
		})
	}
}
