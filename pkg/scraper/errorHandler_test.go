package scraper

import (
	"errors"
	"strings"
	"testing"
)

func TestMultiError_Error(t *testing.T) {
	tests := []struct {
		name   string
		errors map[string]error
		want   []string // Substrings that should be in the error message
	}{
		{
			name:   "empty errors",
			errors: map[string]error{},
			want:   []string{"Multiple errors occurred:"},
		},
		{
			name: "single error",
			errors: map[string]error{
				"example.com": errors.New("connection failed"),
			},
			want: []string{
				"Multiple errors occurred:",
				"Domain: example.com",
				"Error: connection failed",
			},
		},
		{
			name: "multiple errors",
			errors: map[string]error{
				"example.com": errors.New("connection failed"),
				"google.com":  errors.New("timeout"),
			},
			want: []string{
				"Multiple errors occurred:",
				"Domain: example.com",
				"Error: connection failed",
				"Domain: google.com",
				"Error: timeout",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			me := &MultiError{
				Errors: tt.errors,
			}
			got := me.Error()

			// Check that all expected substrings are in the error message
			for _, substr := range tt.want {
				if !strings.Contains(got, substr) {
					t.Errorf("MultiError.Error() = %v, should contain %v", got, substr)
				}
			}
		})
	}
}
