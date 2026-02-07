package logger

import "testing"

func TestRedactPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "path with UUID",
			path:     "/llu/connections/699a9492-1539-11f0-8f17-5e425eb41739/graph",
			expected: "/llu/connections/***/graph",
		},
		{
			name:     "path with uppercase UUID",
			path:     "/llu/connections/699A9492-1539-11F0-8F17-5E425EB41739/graph",
			expected: "/llu/connections/***/graph",
		},
		{
			name:     "path with multiple UUIDs",
			path:     "/users/699a9492-1539-11f0-8f17-5e425eb41739/devices/a1b2c3d4-5678-90ab-cdef-1234567890ab",
			expected: "/users/***/devices/***",
		},
		{
			name:     "path without UUID",
			path:     "/v1/measurements/latest",
			expected: "/v1/measurements/latest",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "",
		},
		{
			name:     "just UUID",
			path:     "699a9492-1539-11f0-8f17-5e425eb41739",
			expected: "***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactPath(tt.path)
			if result != tt.expected {
				t.Errorf("RedactPath(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestRedactSensitive(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "non-empty value",
			value:    "secret-token",
			expected: "***",
		},
		{
			name:     "empty value",
			value:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactSensitive(tt.value)
			if result != tt.expected {
				t.Errorf("RedactSensitive(%q) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestRedactEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "valid email",
			email:    "user@example.com",
			expected: "***@***",
		},
		{
			name:     "empty email",
			email:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactEmail(tt.email)
			if result != tt.expected {
				t.Errorf("RedactEmail(%q) = %q, want %q", tt.email, result, tt.expected)
			}
		})
	}
}
