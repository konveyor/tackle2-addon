package scm

import (
	"testing"
)

func TestSanitizeHost(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		expected string
	}{
		{
			name:     "plain host",
			host:     "proxy.example.com",
			expected: "proxy.example.com",
		},
		{
			name:     "host with http scheme",
			host:     "http://proxy.example.com",
			expected: "proxy.example.com",
		},
		{
			name:     "host with https scheme",
			host:     "https://proxy.example.com",
			expected: "proxy.example.com",
		},
		{
			name:     "host with trailing slash",
			host:     "proxy.example.com/",
			expected: "proxy.example.com",
		},
		{
			name:     "host with http scheme and trailing slash",
			host:     "http://proxy.example.com/",
			expected: "proxy.example.com",
		},
		{
			name:     "host with https scheme and trailing slash",
			host:     "https://proxy.example.com/",
			expected: "proxy.example.com",
		},
		{
			name:     "host with multiple trailing slashes",
			host:     "proxy.example.com///",
			expected: "proxy.example.com",
		},
		{
			name:     "empty host",
			host:     "",
			expected: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizeHost(tc.host)
			if result != tc.expected {
				t.Errorf("sanitizeHost(%q) = %q, want %q", tc.host, result, tc.expected)
			}
		})
	}
}
