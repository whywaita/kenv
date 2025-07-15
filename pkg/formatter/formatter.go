package formatter

import (
	"fmt"
	"strings"

	"github.com/whywaita/kenv/pkg/extractor"
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
		parts = append(parts, fmt.Sprintf("-e %s=%s", env.Name, value))
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
