package extractor

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

// ExtractFromPodSpec extracts environment variables from a PodSpec
func ExtractFromPodSpec(spec *corev1.PodSpec, containerName string) []EnvVar {
	var result []EnvVar

	containers := append(spec.InitContainers, spec.Containers...)
	for _, container := range containers {
		// Skip if container name is specified and doesn't match
		if containerName != "" && container.Name != containerName {
			continue
		}

		// Direct env vars
		for _, env := range container.Env {
			ev := EnvVar{
				Name:  env.Name,
				Value: env.Value,
			}

			// Handle valueFrom
			if env.ValueFrom != nil {
				if env.ValueFrom.SecretKeyRef != nil {
					ev.Source = SourceSecret
					ev.SecretRef = &SecretKeyRef{
						Name: env.ValueFrom.SecretKeyRef.Name,
						Key:  env.ValueFrom.SecretKeyRef.Key,
					}
					if ev.Value == "" {
						ev.Value = fmt.Sprintf("<%s:%s>", ev.SecretRef.Name, ev.SecretRef.Key)
					}
				} else if env.ValueFrom.ConfigMapKeyRef != nil {
					ev.Source = SourceConfigMap
					ev.ConfigRef = &ConfigMapKeyRef{
						Name: env.ValueFrom.ConfigMapKeyRef.Name,
						Key:  env.ValueFrom.ConfigMapKeyRef.Key,
					}
					if ev.Value == "" {
						ev.Value = fmt.Sprintf("<%s:%s>", ev.ConfigRef.Name, ev.ConfigRef.Key)
					}
				}
			} else {
				ev.Source = SourceDirect
			}

			result = append(result, ev)
		}

		// EnvFrom
		for _, envFrom := range container.EnvFrom {
			prefix := ""
			if envFrom.Prefix != "" {
				prefix = envFrom.Prefix
			}

			if envFrom.SecretRef != nil {
				result = append(result, EnvVar{
					Name:   fmt.Sprintf("# from secret: %s", envFrom.SecretRef.Name),
					Value:  "",
					Source: SourceSecret,
					SecretRef: &SecretKeyRef{
						Name: envFrom.SecretRef.Name,
						Key:  "*", // All keys
					},
					Prefix: prefix,
				})
			} else if envFrom.ConfigMapRef != nil {
				result = append(result, EnvVar{
					Name:   fmt.Sprintf("# from configmap: %s", envFrom.ConfigMapRef.Name),
					Value:  "",
					Source: SourceConfigMap,
					ConfigRef: &ConfigMapKeyRef{
						Name: envFrom.ConfigMapRef.Name,
						Key:  "*", // All keys
					},
					Prefix: prefix,
				})
			}
		}

		// Only process first matching container if name was specified
		if containerName != "" {
			break
		}
	}

	return result
}

