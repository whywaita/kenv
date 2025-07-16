package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "0.1.2"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keex",
		Short: "Extract environment variables from Kubernetes manifests",
		Long: `keex is a CLI tool that extracts environment variables from Kubernetes manifests
and formats them for use with docker run or shell commands.`,
		Version: version,
	}

	cmd.AddCommand(newExtractCmd())

	return cmd
}
