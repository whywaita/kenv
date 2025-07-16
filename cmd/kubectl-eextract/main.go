package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/whywaita/kenv/pkg/extractor"
	"github.com/whywaita/kenv/pkg/formatter"
	"github.com/whywaita/kenv/pkg/resolver"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

var (
	version = "0.1.0"
)

type Options struct {
	configFlags *genericclioptions.ConfigFlags
	genericclioptions.IOStreams
}

func main() {
	cmd := NewCmd(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func NewCmd(streams genericclioptions.IOStreams) *cobra.Command {
	o := &Options{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}

	cmd := &cobra.Command{
		Use:   "kubectl-eextract [TYPE/NAME]",
		Short: "Extract environment variables from Kubernetes resources",
		Long: `kubectl-eextract is a kubectl plugin that extracts environment variables from Kubernetes resources
and formats them for use with docker run or shell commands.

Supports Deployment, StatefulSet, DaemonSet, Job, CronJob, and Pod resources.

Examples:
  # Extract env vars from a deployment
  kubectl eextract deployment/my-app

  # Extract env vars from a specific container
  kubectl eextract deployment/my-app -c my-container

  # Output in docker run format
  kubectl eextract deployment/my-app --format docker

  # Output in shell format with export
  kubectl eextract pod/mypod --format shell --export`,
		Version: version,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExtract(o, cmd, args[0])
		},
	}

	o.configFlags.AddFlags(cmd.Flags())

	cmd.Flags().StringP("container", "c", "", "Specify container name (optional)")
	cmd.Flags().StringP("format", "f", "docker", "Output format: docker, shell, dotenv, compose")
	cmd.Flags().BoolP("export", "e", false, "Add export prefix for shell format")

	return cmd
}

func runExtract(o *Options, cmd *cobra.Command, resource string) error {
	restConfig, err := o.configFlags.ToRESTConfig()
	if err != nil {
		return fmt.Errorf("failed to get REST config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	namespace, _, err := o.configFlags.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return fmt.Errorf("failed to get namespace: %w", err)
	}

	// Parse resource type and name
	parts := strings.SplitN(resource, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid resource format, expected TYPE/NAME")
	}
	resourceType, resourceName := parts[0], parts[1]

	// Extract based on resource type
	var envVars []extractor.EnvVar
	ctx := context.Background()

	switch strings.ToLower(resourceType) {
	case "deployment", "deploy":
		deploy, err := clientset.AppsV1().Deployments(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get deployment: %w", err)
		}
		envVars = extractFromPodSpec(&deploy.Spec.Template.Spec, cmd.Flag("container").Value.String())

	case "statefulset", "sts":
		sts, err := clientset.AppsV1().StatefulSets(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get statefulset: %w", err)
		}
		envVars = extractFromPodSpec(&sts.Spec.Template.Spec, cmd.Flag("container").Value.String())

	case "daemonset", "ds":
		ds, err := clientset.AppsV1().DaemonSets(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get daemonset: %w", err)
		}
		envVars = extractFromPodSpec(&ds.Spec.Template.Spec, cmd.Flag("container").Value.String())

	case "pod", "po":
		pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get pod: %w", err)
		}
		envVars = extractFromPodSpec(&pod.Spec, cmd.Flag("container").Value.String())

	case "job":
		job, err := clientset.BatchV1().Jobs(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get job: %w", err)
		}
		envVars = extractFromPodSpec(&job.Spec.Template.Spec, cmd.Flag("container").Value.String())

	case "cronjob", "cj":
		cj, err := clientset.BatchV1().CronJobs(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get cronjob: %w", err)
		}
		envVars = extractFromPodSpec(&cj.Spec.JobTemplate.Spec.Template.Spec, cmd.Flag("container").Value.String())

	default:
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	// Resolve secrets and configmaps
	res := resolver.NewFromClientset(clientset, namespace)
	envVars, err = res.ResolveAll(envVars)
	if err != nil {
		fmt.Fprintf(o.IOStreams.ErrOut, "Warning: failed to resolve some references: %v\n", err)
	}

	// Format output
	formatFlag, _ := cmd.Flags().GetString("format")
	exportFlag, _ := cmd.Flags().GetBool("export")

	var output string
	switch formatFlag {
	case "docker":
		output = formatter.FormatDocker(envVars, false)
	case "shell":
		output = formatter.FormatShell(envVars, exportFlag)
	case "dotenv":
		output = formatter.FormatDotenv(envVars)
	case "compose":
		output = formatter.FormatCompose(envVars)
	default:
		output = formatter.FormatDocker(envVars, false)
	}

	fmt.Fprintln(o.IOStreams.Out, output)
	return nil
}

func extractFromPodSpec(spec *corev1.PodSpec, containerName string) []extractor.EnvVar {
	var result []extractor.EnvVar

	containers := append(spec.InitContainers, spec.Containers...)
	for _, container := range containers {
		// Skip if container name is specified and doesn't match
		if containerName != "" && container.Name != containerName {
			continue
		}

		// Direct env vars
		for _, env := range container.Env {
			ev := extractor.EnvVar{
				Name:  env.Name,
				Value: env.Value,
			}

			// Handle valueFrom
			if env.ValueFrom != nil {
				if env.ValueFrom.SecretKeyRef != nil {
					ev.SecretRef = &extractor.SecretKeyRef{
						Name: env.ValueFrom.SecretKeyRef.Name,
						Key:  env.ValueFrom.SecretKeyRef.Key,
					}
				} else if env.ValueFrom.ConfigMapKeyRef != nil {
					ev.ConfigRef = &extractor.ConfigMapKeyRef{
						Name: env.ValueFrom.ConfigMapKeyRef.Name,
						Key:  env.ValueFrom.ConfigMapKeyRef.Key,
					}
				}
			}

			result = append(result, ev)
		}

		// EnvFrom
		for _, envFrom := range container.EnvFrom {
			if envFrom.SecretRef != nil {
				result = append(result, extractor.EnvVar{
					Name:  fmt.Sprintf("# from secret: %s", envFrom.SecretRef.Name),
					Value: "",
					SecretRef: &extractor.SecretKeyRef{
						Name: envFrom.SecretRef.Name,
						Key:  "*", // All keys
					},
				})
			} else if envFrom.ConfigMapRef != nil {
				result = append(result, extractor.EnvVar{
					Name:  fmt.Sprintf("# from configmap: %s", envFrom.ConfigMapRef.Name),
					Value: "",
					ConfigRef: &extractor.ConfigMapKeyRef{
						Name: envFrom.ConfigMapRef.Name,
						Key:  "*", // All keys
					},
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
