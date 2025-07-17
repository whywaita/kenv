package formatter

import (
	"testing"

	"github.com/whywaita/keex/pkg/extractor"
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

func TestFormatShellOneLine(t *testing.T) {
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
			redact: false,
			expected: `FOO='bar'
BAZ='qux'`,
		},
		{
			name: "with single quotes",
			envVars: []extractor.EnvVar{
				{Name: "MESSAGE", Value: `hello 'world'`, Source: extractor.SourceDirect},
			},
			redact:   false,
			expected: `MESSAGE='hello '\''world'\'''`,
		},
		{
			name: "redacted secrets",
			envVars: []extractor.EnvVar{
				{Name: "PASSWORD", Value: "secret123", Source: extractor.SourceSecret, IsSecret: true},
				{Name: "USER", Value: "admin", Source: extractor.SourceDirect},
			},
			redact: true,
			expected: `PASSWORD='***REDACTED***'
USER='admin'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatShell(tt.envVars, false, tt.redact)
			if result != tt.expected {
				t.Errorf("FormatShell() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatDotenv(t *testing.T) {
	tests := []struct {
		name     string
		envVars  []extractor.EnvVar
		expected string
	}{
		{
			name: "simple env vars",
			envVars: []extractor.EnvVar{
				{Name: "FOO", Value: "bar", Source: extractor.SourceDirect},
				{Name: "BAZ", Value: "qux", Source: extractor.SourceDirect},
			},
			expected: `FOO="bar"
BAZ="qux"`,
		},
		{
			name: "with quotes",
			envVars: []extractor.EnvVar{
				{Name: "MESSAGE", Value: `hello "world"`, Source: extractor.SourceDirect},
			},
			expected: `MESSAGE="hello \"world\""`,
		},
		{
			name: "skip comments",
			envVars: []extractor.EnvVar{
				{Name: "# This is a comment", Value: "ignored", Source: extractor.SourceDirect},
				{Name: "FOO", Value: "bar", Source: extractor.SourceDirect},
			},
			expected: `FOO="bar"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDotenv(tt.envVars)
			if result != tt.expected {
				t.Errorf("FormatDotenv() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatCompose(t *testing.T) {
	tests := []struct {
		name     string
		envVars  []extractor.EnvVar
		expected string
	}{
		{
			name: "simple env vars",
			envVars: []extractor.EnvVar{
				{Name: "FOO", Value: "bar", Source: extractor.SourceDirect},
				{Name: "BAZ", Value: "qux", Source: extractor.SourceDirect},
			},
			expected: `environment:
  FOO: "bar"
  BAZ: "qux"`,
		},
		{
			name: "multiline value",
			envVars: []extractor.EnvVar{
				{Name: "SCRIPT", Value: "line1\nline2\nline3", Source: extractor.SourceDirect},
			},
			expected: `environment:
  SCRIPT: |
    line1
    line2
    line3`,
		},
		{
			name: "skip comments",
			envVars: []extractor.EnvVar{
				{Name: "# This is a comment", Value: "ignored", Source: extractor.SourceDirect},
				{Name: "FOO", Value: "bar", Source: extractor.SourceDirect},
			},
			expected: `environment:
  FOO: "bar"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCompose(tt.envVars)
			if result != tt.expected {
				t.Errorf("FormatCompose() = %q, want %q", result, tt.expected)
			}
		})
	}
}
