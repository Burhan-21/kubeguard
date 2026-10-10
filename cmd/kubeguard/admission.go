package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/Burhan-21/kubeguard/internal/admission"
	"github.com/Burhan-21/kubeguard/internal/policy"
	"github.com/Burhan-21/kubeguard/internal/rules/reliability"
	"github.com/Burhan-21/kubeguard/internal/rules/security"
)

var (
	admissionPort            int
	admissionCertFile        string
	admissionKeyFile         string
	admissionMode            string
	admissionProfile         string
	admissionSecretName      string
	admissionSecretNamespace string
	admissionReloadInterval  time.Duration
	certAlias                string
	keyAlias                 string
	profileAlias             string
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

		var kubeClient kubernetes.Interface
		if admissionSecretName != "" {
			if admissionSecretNamespace == "" {
				if podNs := os.Getenv("POD_NAMESPACE"); podNs != "" {
					admissionSecretNamespace = podNs
				} else {
					admissionSecretNamespace = "kubeguard-system"
				}
			}

			if cfg, err := rest.InClusterConfig(); err == nil {
				if clientset, err := kubernetes.NewForConfig(cfg); err == nil {
					kubeClient = clientset
					fmt.Fprintf(os.Stderr, "Configured Kubernetes Secret watcher for %s/%s\n", admissionSecretNamespace, admissionSecretName)
				} else {
					fmt.Fprintf(os.Stderr, "Warning: Failed to create in-cluster Kubernetes client: %v; falling back to file-based TLS rotation\n", err)
				}
			} else {
				fmt.Fprintf(os.Stderr, "In-cluster Kubernetes client not available: %v; falling back to file-based TLS rotation\n", err)
			}
		}

		server := &admission.Server{
			Handler:         handler,
			CertFile:        admissionCertFile,
			KeyFile:         admissionKeyFile,
			Port:            admissionPort,
			SecretName:      admissionSecretName,
			SecretNamespace: admissionSecretNamespace,
			KubeClient:      kubeClient,
			ReloadInterval:  admissionReloadInterval,
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
	admissionCmd.Flags().StringVar(&admissionSecretName, "tls-secret-name", "", "Name of Kubernetes TLS Secret to watch for dynamic rotation")
	admissionCmd.Flags().StringVar(&admissionSecretNamespace, "tls-secret-namespace", "", "Namespace of Kubernetes TLS Secret to watch (defaults to POD_NAMESPACE or kubeguard-system)")
	admissionCmd.Flags().DurationVar(&admissionReloadInterval, "tls-reload-interval", 1*time.Second, "Polling interval for checking TLS certificate file updates")
	admissionCmd.Flags().StringVar(&admissionMode, "mode", "enforce", "Webhook evaluation mode (audit, warn, enforce)")
	admissionCmd.Flags().StringVarP(&admissionProfile, "profile", "p", "", "Path to policy profile YAML file")

	// Compatibility aliases for flags
	admissionCmd.Flags().StringVar(&certAlias, "tls-cert-file", "", "Path to TLS certificate file")
	admissionCmd.Flags().StringVar(&keyAlias, "tls-private-key-file", "", "Path to TLS private key file")
	admissionCmd.Flags().StringVar(&profileAlias, "policy-profile", "", "Path to policy profile YAML file")
}
