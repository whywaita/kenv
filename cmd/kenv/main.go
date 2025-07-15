package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kenv",
		Short: "Extract environment variables from Kubernetes manifests",
		Long: `kenv is a CLI tool that extracts environment variables from Kubernetes manifests
and formats them for use with docker run or shell commands.`,
		Version: version,
	}

	cmd.AddCommand(newExtractCmd())

	return cmd
}

