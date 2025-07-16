package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/whywaita/keex/pkg/extractor"
	"github.com/whywaita/keex/pkg/formatter"
	"github.com/whywaita/keex/pkg/resolver"
)

type extractOptions struct {
	file      string
	mode      string
	container string
	context   string
	namespace string
	redact    bool
}

func newExtractCmd() *cobra.Command {
	opts := &extractOptions{}

	cmd := &cobra.Command{
		Use:   "extract",
		Short: "Extract environment variables from Kubernetes manifests",
		Long: `Extract environment variables from Kubernetes manifests and format them
for use with docker run or shell commands.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExtract(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.file, "file", "f", "", "Manifest file path (\"-\" for stdin)")
	cmd.Flags().StringVar(&opts.mode, "mode", "env", "Output mode: docker|env")
	cmd.Flags().StringVar(&opts.container, "container", "", "Target container name")
	cmd.Flags().StringVar(&opts.context, "context", "", "kubeconfig context (default: current)")
	cmd.Flags().StringVar(&opts.namespace, "namespace", "", "Kubernetes namespace (default: manifest/ns)")
	cmd.Flags().BoolVar(&opts.redact, "redact", false, "Mask secret values in output")

	return cmd
}

func runExtract(opts *extractOptions) error {
	// Validate mode
	if opts.mode != "docker" && opts.mode != "env" {
		return fmt.Errorf("invalid mode: %s (must be docker or env)", opts.mode)
	}

	// Read manifest
	var reader io.Reader
	switch opts.file {
	case "":
		return fmt.Errorf("manifest file is required")
	case "-":
		reader = os.Stdin
	default:
		file, err := os.Open(opts.file)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "failed to close file: %v\n", err)
			}
		}()
		reader = file
	}

	// Extract environment variables
	ext := extractor.New()
	envVars, err := ext.Extract(reader, extractor.Options{
		Container: opts.container,
	})
	if err != nil {
		return fmt.Errorf("failed to extract environment variables: %w", err)
	}

	// Try to resolve secrets/configmaps if kubeconfig is available
	res, err := resolver.New(resolver.Options{
		Context:   opts.context,
		Namespace: opts.namespace,
	})
	if err == nil {
		// Kubeconfig is available, resolve secrets/configmaps
		envVars, err = res.ResolveAll(envVars)
		if err != nil {
			return fmt.Errorf("failed to resolve secrets: %w", err)
		}
	}
	// If kubeconfig is not available, just continue with placeholder values

	// Format output
	var output string
	switch opts.mode {
	case "docker":
		output = formatter.FormatDocker(envVars, opts.redact)
	case "env":
		output = formatter.FormatEnv(envVars, opts.redact)
	}

	fmt.Println(output)
	return nil
}
