package formatter

import (
	"testing"

	"github.com/whywaita/kenv/pkg/extractor"
)

func TestFormatDocker(t *testing.T) {
	tests := []struct {
		name     string
		envVars  []extractor.EnvVar
		redact   bool
		expected string
	}{
		{
			name: "simple env vars",
			envVars: []extractor.EnvVar{
				{Name: "FOO", Value: "bar", Source: extractor.SourceDirect},
				{Name: "BAZ", Value: "qux", Source: extractor.SourceDirect},
			},
			redact:   false,
			expected: `-e FOO="bar" -e BAZ="qux"`,
		},
		{
			name: "with quotes",
			envVars: []extractor.EnvVar{
				{Name: "MESSAGE", Value: `hello "world"`, Source: extractor.SourceDirect},
			},
			redact:   false,
			expected: `-e MESSAGE="hello \"world\""`,
		},
		{
			name: "redacted secrets",
			envVars: []extractor.EnvVar{
				{Name: "PASSWORD", Value: "secret123", Source: extractor.SourceSecret, IsSecret: true},
				{Name: "USER", Value: "admin", Source: extractor.SourceDirect},
			},
			redact:   true,
			expected: `-e PASSWORD="***REDACTED***" -e USER="admin"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDocker(tt.envVars, tt.redact)
			if result != tt.expected {
				t.Errorf("FormatDocker() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatEnv(t *testing.T) {
	tests := []struct {
		name     string
		envVars  []extractor.EnvVar
		redact   bool
		expected string
	}{
		{
			name: "simple env vars",
			envVars: []extractor.EnvVar{
				{Name: "FOO", Value: "bar", Source: extractor.SourceDirect},
				{Name: "BAZ", Value: "qux", Source: extractor.SourceDirect},
			},
			redact:   false,
			expected: `FOO="bar" BAZ="qux"`,
		},
		{
			name: "with quotes",
			envVars: []extractor.EnvVar{
				{Name: "MESSAGE", Value: `hello "world"`, Source: extractor.SourceDirect},
			},
			redact:   false,
			expected: `MESSAGE="hello \"world\""`,
		},
		{
			name: "redacted secrets",
			envVars: []extractor.EnvVar{
				{Name: "PASSWORD", Value: "secret123", Source: extractor.SourceSecret, IsSecret: true},
				{Name: "USER", Value: "admin", Source: extractor.SourceDirect},
			},
			redact:   true,
			expected: `PASSWORD="***REDACTED***" USER="admin"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatEnv(tt.envVars, tt.redact)
			if result != tt.expected {
				t.Errorf("FormatEnv() = %q, want %q", result, tt.expected)
			}
		})
	}
}
