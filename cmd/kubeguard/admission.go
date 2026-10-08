package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Burhan-21/kubeguard/internal/admission"
	"github.com/Burhan-21/kubeguard/internal/policy"
	"github.com/Burhan-21/kubeguard/internal/rules/reliability"
	"github.com/Burhan-21/kubeguard/internal/rules/security"
)

var (
	admissionPort     int
	admissionCertFile string
	admissionKeyFile  string
	admissionMode     string
	admissionProfile  string
	certAlias         string
	keyAlias          string
	profileAlias      string
)

var admissionCmd = &cobra.Command{
	Use:     "admission",
	Aliases: []string{"webhook"},
	Short:   "Start KubeGuard Validating Admission Webhook server",
	Long:    `Runs the KubeGuard HTTPS webhook server to intercept admission requests and evaluate them against policy rules.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if admissionCertFile == "" && certAlias != "" {
			admissionCertFile = certAlias
		}
		if admissionKeyFile == "" && keyAlias != "" {
			admissionKeyFile = keyAlias
		}
		if admissionProfile == "" && profileAlias != "" {
			admissionProfile = profileAlias
		}

		// Collect all rules
		var allRules []policy.Rule
		for _, r := range security.AllSecurityRules() {
			allRules = append(allRules, r)
		}
		for _, r := range reliability.AllReliabilityRules() {
			allRules = append(allRules, r)
		}

		engine := policy.NewEngine(allRules)

		// Apply profile overrides if specified
		if admissionProfile != "" {
			prof, err := policy.LoadProfile(admissionProfile)
			if err != nil {
				return fmt.Errorf("failed to load profile %s: %w", admissionProfile, err)
			}
			disabled, overrides := prof.BuildOverrides()
			engine.SetDisabled(disabled)
			engine.SetOverrides(overrides)
		}

		handler := &admission.Handler{
			Engine: engine,
			Mode:   admissionMode,
		}

		server := &admission.Server{
			Handler:  handler,
			CertFile: admissionCertFile,
			KeyFile:  admissionKeyFile,
			Port:     admissionPort,
		}

		fmt.Fprintf(os.Stderr, "Starting KubeGuard admission webhook server on port %d (mode: %s)...\n", admissionPort, admissionMode)
		return server.Start()
	},
}

func init() {
	rootCmd.AddCommand(admissionCmd)
	admissionCmd.Flags().IntVar(&admissionPort, "port", 8443, "Webhook listening port")
	admissionCmd.Flags().StringVar(&admissionCertFile, "tls-cert", "", "Path to TLS certificate file")
	admissionCmd.Flags().StringVar(&admissionKeyFile, "tls-key", "", "Path to TLS private key file")
	admissionCmd.Flags().StringVar(&admissionMode, "mode", "enforce", "Webhook evaluation mode (audit, warn, enforce)")
	admissionCmd.Flags().StringVarP(&admissionProfile, "profile", "p", "", "Path to policy profile YAML file")

	// Compatibility aliases for flags
	admissionCmd.Flags().StringVar(&certAlias, "tls-cert-file", "", "Path to TLS certificate file")
	admissionCmd.Flags().StringVar(&keyAlias, "tls-private-key-file", "", "Path to TLS private key file")
	admissionCmd.Flags().StringVar(&profileAlias, "policy-profile", "", "Path to policy profile YAML file")
}
