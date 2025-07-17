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

		if len(rawObj.Raw) == 0 {
			continue
		}

		obj, gvk, err := e.decoder.Decode(rawObj.Raw, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to decode object: %w", err)
		}

		var extractedVars []EnvVar

		switch gvk.Kind {
		case "Deployment":
			deployment := obj.(*appsv1.Deployment)
			extractedVars = ExtractFromPodSpec(&deployment.Spec.Template.Spec, opts.Container)
		case "StatefulSet":
			statefulSet := obj.(*appsv1.StatefulSet)
			extractedVars = ExtractFromPodSpec(&statefulSet.Spec.Template.Spec, opts.Container)
		case "DaemonSet":
			daemonSet := obj.(*appsv1.DaemonSet)
			extractedVars = ExtractFromPodSpec(&daemonSet.Spec.Template.Spec, opts.Container)
		case "Job":
			job := obj.(*batchv1.Job)
			extractedVars = ExtractFromPodSpec(&job.Spec.Template.Spec, opts.Container)
		case "CronJob":
			cronJob := obj.(*batchv1.CronJob)
			extractedVars = ExtractFromPodSpec(&cronJob.Spec.JobTemplate.Spec.Template.Spec, opts.Container)
		case "Pod":
			pod := obj.(*corev1.Pod)
			extractedVars = ExtractFromPodSpec(&pod.Spec, opts.Container)
		default:
			return nil, fmt.Errorf("unsupported resource type: %s", gvk.Kind)
		}

		// Mark secrets as IsSecret for redaction support in keex
		for i := range extractedVars {
			if extractedVars[i].Source == SourceSecret {
				extractedVars[i].IsSecret = true
			}
		}

		envVars = append(envVars, extractedVars...)
	}

	if len(envVars) == 0 {
		return nil, fmt.Errorf("no environment variables found")
	}

	return envVars, nil
}
