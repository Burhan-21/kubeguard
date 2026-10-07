package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
	"github.com/Burhan-21/kubeguard/internal/parser"
	"github.com/Burhan-21/kubeguard/internal/policy"
	"github.com/Burhan-21/kubeguard/internal/reporter"
	"github.com/Burhan-21/kubeguard/internal/rules/reliability"
	"github.com/Burhan-21/kubeguard/internal/rules/security"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var scanCmd = &cobra.Command{
	Use:   "scan [flags] [files...]",
	Short: "Scan Kubernetes manifests for security and reliability issues",
	Long:  `Scan one or more Kubernetes YAML manifests or directories against KubeGuard's security and reliability rules.`,
	RunE:  runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().StringP("profile", "p", "", "Path to policy profile YAML file")
	scanCmd.Flags().StringP("output", "o", "human", "Output format (human, json, sarif)")
	scanCmd.Flags().String("fail-on", "block", "Exit non-zero on severity level (warn, block)")
}

func runScan(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no files provided to scan")
	}

	outputFormat, _ := cmd.Flags().GetString("output")
	failOn, _ := cmd.Flags().GetString("fail-on")
	profilePath, _ := cmd.Flags().GetString("profile")

	// Parse all inputs
	var allResources []unstructured.Unstructured
	for _, path := range args {
		var resources []unstructured.Unstructured
		var err error

		if path == "-" {
			resources, err = parser.ParseReader(os.Stdin)
		} else {
			fi, statErr := os.Stat(path)
			if statErr != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", statErr)
				continue
			}
			if fi.IsDir() {
				resources, err = parser.ParseDirectory(path)
			} else {
				resources, err = parser.Parse(path)
			}
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", path, err)
			continue
		}
		allResources = append(allResources, resources...)
	}

	if len(allResources) == 0 {
		fmt.Fprintln(os.Stderr, "No Kubernetes resources found")
		os.Exit(3)
	}

	// Normalize resources
	var normalized []*normalizer.NormalizedResource
	for _, u := range allResources {
		nr, err := normalizer.Normalize(u)
		if err != nil {
			// Unsupported kinds are logged but not fatal
			fmt.Fprintf(os.Stderr, "Skipping %s/%s: %v\n", u.GetKind(), u.GetName(), err)
			continue
		}
		normalized = append(normalized, nr)
	}

	// Build rule set
	var allRules []policy.Rule
	for _, r := range security.AllSecurityRules() {
		allRules = append(allRules, r)
	}
	for _, r := range reliability.AllReliabilityRules() {
		allRules = append(allRules, r)
	}

	// Create engine
	engine := policy.NewEngine(allRules)

	// Apply profile overrides
	if profilePath != "" {
		prof, err := policy.LoadProfile(profilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading profile: %v\n", err)
			os.Exit(3)
		}
		disabled, overrides := prof.BuildOverrides()
		engine.SetDisabled(disabled)
		engine.SetOverrides(overrides)
	}

	// Evaluate
	result := engine.EvaluateAll(normalized)

	// Report
	if err := reporter.Report(outputFormat, result, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error reporting: %v\n", err)
		os.Exit(3)
	}

	// Exit code based on fail-on level
	switch strings.ToLower(failOn) {
	case "warn":
		if result.OverallResult == findings.SeverityWarn || result.OverallResult == findings.SeverityBlock {
			os.Exit(result.ExitCode())
		}
	case "block":
		if result.OverallResult == findings.SeverityBlock {
			os.Exit(2)
		}
	}

	os.Exit(0)
	return nil
}
