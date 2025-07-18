package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/whywaita/keex/pkg/extractor"
	"github.com/whywaita/keex/pkg/formatter"
	"github.com/whywaita/keex/pkg/resolver"
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
		Use:   "kubectl-eex TYPE NAME or kubectl-eex TYPE/NAME",
		Short: "Extract environment variables from Kubernetes resources",
		Long: `kubectl-eex is a kubectl plugin that extracts environment variables from Kubernetes resources
and formats them for use with docker run or shell commands.

Supports Deployment, StatefulSet, DaemonSet, Job, CronJob, and Pod resources.

Examples:
  # Extract env vars from a deployment (both formats supported)
  kubectl eex deployment/my-app
  kubectl eex deployment my-app

  # Extract env vars from a specific container
  kubectl eex deployment/my-app -c my-container
  kubectl eex deployment my-app -c my-container

  # Output in docker run format
  kubectl eex deployment/my-app --format docker

  # Output in shell format with export
  kubectl eex pod/mypod --format shell --export`,
		Version: version,
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExtract(o, cmd, args)
		},
	}

	o.configFlags.AddFlags(cmd.Flags())

	cmd.Flags().StringP("container", "c", "", "Specify container name (optional)")
	cmd.Flags().StringP("format", "f", "docker", "Output format: docker, shell, dotenv, compose")
	cmd.Flags().BoolP("export", "e", false, "Add export prefix for shell format")

	return cmd
}

func runExtract(o *Options, cmd *cobra.Command, args []string) error {
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
	var resourceType, resourceName string

	if len(args) == 1 {
		// Handle TYPE/NAME format
		parts := strings.SplitN(args[0], "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid resource format, expected TYPE/NAME or TYPE NAME")
		}
		resourceType, resourceName = parts[0], parts[1]
	} else if len(args) == 2 {
		// Handle TYPE NAME format
		resourceType, resourceName = args[0], args[1]
	} else {
		return fmt.Errorf("invalid arguments, expected TYPE/NAME or TYPE NAME")
	}

	// Extract based on resource type
	var envVars []extractor.EnvVar
	ctx := context.Background()

	switch strings.ToLower(resourceType) {
	case "deployment", "deploy":
		deploy, err := clientset.AppsV1().Deployments(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get deployment: %w", err)
		}
		envVars = extractor.ExtractFromPodSpec(&deploy.Spec.Template.Spec, cmd.Flag("container").Value.String())

	case "statefulset", "sts":
		sts, err := clientset.AppsV1().StatefulSets(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get statefulset: %w", err)
		}
		envVars = extractor.ExtractFromPodSpec(&sts.Spec.Template.Spec, cmd.Flag("container").Value.String())

	case "daemonset", "ds":
		ds, err := clientset.AppsV1().DaemonSets(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get daemonset: %w", err)
		}
		envVars = extractor.ExtractFromPodSpec(&ds.Spec.Template.Spec, cmd.Flag("container").Value.String())

	case "pod", "po":
		pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get pod: %w", err)
		}
		envVars = extractor.ExtractFromPodSpec(&pod.Spec, cmd.Flag("container").Value.String())

	case "job":
		job, err := clientset.BatchV1().Jobs(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get job: %w", err)
		}
		envVars = extractor.ExtractFromPodSpec(&job.Spec.Template.Spec, cmd.Flag("container").Value.String())

	case "cronjob", "cj":
		cj, err := clientset.BatchV1().CronJobs(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get cronjob: %w", err)
		}
		envVars = extractor.ExtractFromPodSpec(&cj.Spec.JobTemplate.Spec.Template.Spec, cmd.Flag("container").Value.String())

	default:
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	// Resolve secrets and configmaps
	res := resolver.NewFromClientset(clientset, namespace)
	envVars, err = res.ResolveAll(envVars)
	if err != nil {
		if _, writeErr := fmt.Fprintf(o.ErrOut, "Warning: failed to resolve some references: %v\n", err); writeErr != nil {
			return writeErr
		}
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

	if _, err := fmt.Fprintln(o.Out, output); err != nil {
		return err
	}
	return nil
}
