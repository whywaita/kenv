package formatter

import (
	"fmt"
	"strings"

	"github.com/whywaita/keex/pkg/extractor"
)

func FormatDocker(envVars []extractor.EnvVar, redact bool) string {
	var parts []string

	for _, env := range envVars {
		value := env.Value
		if redact && env.IsSecret {
			value = "***REDACTED***"
		}
		// Escape double quotes in value for docker
		value = strings.ReplaceAll(value, `"`, `\"`)
		parts = append(parts, fmt.Sprintf(`-e %s="%s"`, env.Name, value))
	}

	return strings.Join(parts, " ")
}

func FormatEnv(envVars []extractor.EnvVar, redact bool) string {
	var parts []string

	for _, env := range envVars {
		value := env.Value
		if redact && env.IsSecret {
			value = "***REDACTED***"
		}
		// Escape double quotes in value for shell
		value = strings.ReplaceAll(value, `"`, `\"`)
		parts = append(parts, fmt.Sprintf(`%s="%s"`, env.Name, value))
	}

	return strings.Join(parts, " ")
}

func FormatShell(envVars []extractor.EnvVar, export bool) string {
	var parts []string

	for _, env := range envVars {
		// Skip comment entries
		if strings.HasPrefix(env.Name, "#") {
			continue
		}

		value := env.Value
		// Escape single quotes in value for shell
		value = strings.ReplaceAll(value, `'`, `'\''`)

		if export {
			parts = append(parts, fmt.Sprintf(`export %s='%s'`, env.Name, value))
		} else {
			parts = append(parts, fmt.Sprintf(`%s='%s'`, env.Name, value))
		}
	}

	return strings.Join(parts, "\n")
}

func FormatDotenv(envVars []extractor.EnvVar) string {
	var parts []string

	for _, env := range envVars {
		// Skip comment entries
		if strings.HasPrefix(env.Name, "#") {
			continue
		}

		value := env.Value
		// Escape double quotes in value for dotenv
		value = strings.ReplaceAll(value, `"`, `\"`)
		parts = append(parts, fmt.Sprintf(`%s="%s"`, env.Name, value))
	}

	return strings.Join(parts, "\n")
}

func FormatCompose(envVars []extractor.EnvVar) string {
	var parts []string
	parts = append(parts, "environment:")

	for _, env := range envVars {
		// Skip comment entries
		if strings.HasPrefix(env.Name, "#") {
			continue
		}

		value := env.Value
		// For compose format, we need to handle multi-line values
		if strings.Contains(value, "\n") {
			// Use literal style for multi-line values
			parts = append(parts, fmt.Sprintf("  %s: |", env.Name))
			lines := strings.Split(value, "\n")
			for _, line := range lines {
				parts = append(parts, fmt.Sprintf("    %s", line))
			}
		} else {
			// Single line values with quotes
			value = strings.ReplaceAll(value, `"`, `\"`)
			parts = append(parts, fmt.Sprintf(`  %s: "%s"`, env.Name, value))
		}
	}

	return strings.Join(parts, "\n")
}
