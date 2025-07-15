package extractor

import (
	"fmt"
	"io"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

type Extractor struct {
	decoder runtime.Decoder
}

func New() *Extractor {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = batchv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	return &Extractor{
		decoder: serializer.NewCodecFactory(scheme).UniversalDeserializer(),
	}
}

func (e *Extractor) Extract(reader io.Reader, opts Options) ([]EnvVar, error) {
	yamlReader := utilyaml.NewYAMLOrJSONDecoder(reader, 4096)

	var envVars []EnvVar

	for {
		var rawObj runtime.RawExtension
		if err := yamlReader.Decode(&rawObj); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to decode manifest: %w", err)
		}

		if rawObj.Raw == nil || len(rawObj.Raw) == 0 {
			continue
		}

		obj, gvk, err := e.decoder.Decode(rawObj.Raw, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to decode object: %w", err)
		}

		var containers []corev1.Container

		switch gvk.Kind {
		case "Deployment":
			deployment := obj.(*appsv1.Deployment)
			containers = deployment.Spec.Template.Spec.Containers
		case "StatefulSet":
			statefulSet := obj.(*appsv1.StatefulSet)
			containers = statefulSet.Spec.Template.Spec.Containers
		case "DaemonSet":
			daemonSet := obj.(*appsv1.DaemonSet)
			containers = daemonSet.Spec.Template.Spec.Containers
		case "Job":
			job := obj.(*batchv1.Job)
			containers = job.Spec.Template.Spec.Containers
		case "CronJob":
			cronJob := obj.(*batchv1.CronJob)
			containers = cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers
		case "Pod":
			pod := obj.(*corev1.Pod)
			containers = pod.Spec.Containers
		default:
			return nil, fmt.Errorf("unsupported resource type: %s", gvk.Kind)
		}

		for _, container := range containers {
			if opts.Container != "" && container.Name != opts.Container {
				continue
			}

			// Extract env vars
			for _, env := range container.Env {
				envVar := EnvVar{
					Name: env.Name,
				}

				if env.Value != "" {
					envVar.Value = env.Value
					envVar.Source = SourceDirect
				} else if env.ValueFrom != nil {
					if env.ValueFrom.SecretKeyRef != nil {
						envVar.Source = SourceSecret
						envVar.IsSecret = true
						envVar.SecretRef = &SecretKeyRef{
							Name: env.ValueFrom.SecretKeyRef.Name,
							Key:  env.ValueFrom.SecretKeyRef.Key,
						}
						envVar.Value = fmt.Sprintf("<%s:%s>", env.ValueFrom.SecretKeyRef.Name, env.ValueFrom.SecretKeyRef.Key)
					} else if env.ValueFrom.ConfigMapKeyRef != nil {
						envVar.Source = SourceConfigMap
						envVar.ConfigRef = &ConfigMapKeyRef{
							Name: env.ValueFrom.ConfigMapKeyRef.Name,
							Key:  env.ValueFrom.ConfigMapKeyRef.Key,
						}
						envVar.Value = fmt.Sprintf("<%s:%s>", env.ValueFrom.ConfigMapKeyRef.Name, env.ValueFrom.ConfigMapKeyRef.Key)
					}
				}

				envVars = append(envVars, envVar)
			}

			// Extract envFrom
			for _, envFrom := range container.EnvFrom {
				if envFrom.SecretRef != nil {
					// For envFrom, we'll need to fetch all keys from the secret
					// For now, we'll add a placeholder
					envVar := EnvVar{
						Name:     fmt.Sprintf("FROM_SECRET_%s", envFrom.SecretRef.Name),
						Value:    fmt.Sprintf("<all-from-secret:%s>", envFrom.SecretRef.Name),
						Source:   SourceSecret,
						IsSecret: true,
					}
					envVars = append(envVars, envVar)
				} else if envFrom.ConfigMapRef != nil {
					envVar := EnvVar{
						Name:   fmt.Sprintf("FROM_CONFIGMAP_%s", envFrom.ConfigMapRef.Name),
						Value:  fmt.Sprintf("<all-from-configmap:%s>", envFrom.ConfigMapRef.Name),
						Source: SourceConfigMap,
					}
					envVars = append(envVars, envVar)
				}
			}

			// If container specified, break after first match
			if opts.Container != "" {
				break
			}
		}
	}

	if len(envVars) == 0 {
		return nil, fmt.Errorf("no environment variables found")
	}

	return envVars, nil
}