package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kubeguard",
	Short: "KubeGuard - Kubernetes Deployment Security & Reliability Policy Engine",
	Long:  `KubeGuard is a static analysis tool and admission controller for Kubernetes resource manifests, ensuring security and reliability policies are met.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
