package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Burhan-21/kubeguard/test/load"
	"github.com/spf13/cobra"
)

var (
	loadTargetURL   string
	loadWorkload    string
	loadConcurrency int
	loadRequests    int
	loadQPS         int
	loadTimeout     time.Duration
	loadInsecureTLS bool
	loadOutputJSON  bool
)

var loadCmd = &cobra.Command{
	Use:   "load-test",
	Short: "Execute high-throughput admission load generation against a webhook endpoint",
	Long: `Generate multi-phase synthetic AdmissionReview load (warm-up, sustained, burst, recovery)
against a running KubeGuard admission webhook endpoint and output latency percentiles and error metrics.`,
	RunE: runLoadTest,
}

func init() {
	loadCmd.Flags().StringVar(&loadTargetURL, "target-url", "https://127.0.0.1:8443/validate", "Target admission webhook URL")
	loadCmd.Flags().StringVar(&loadWorkload, "workload", "mixed", "Workload type: compliant, denied, or mixed")
	loadCmd.Flags().IntVar(&loadConcurrency, "concurrency", 10, "Concurrent client workers")
	loadCmd.Flags().IntVar(&loadRequests, "requests", 200, "Total AdmissionReview requests")
	loadCmd.Flags().IntVar(&loadQPS, "qps", 0, "Target QPS limit (0 = max throughput)")
	loadCmd.Flags().DurationVar(&loadTimeout, "timeout", 5*time.Second, "HTTP client timeout per request")
	loadCmd.Flags().BoolVar(&loadInsecureTLS, "insecure-tls", true, "Skip TLS certificate verification for self-signed webhook certs")
	loadCmd.Flags().BoolVar(&loadOutputJSON, "json", false, "Output results in JSON format")

	rootCmd.AddCommand(loadCmd)
}

func runLoadTest(cmd *cobra.Command, args []string) error {
	var wt load.WorkloadType
	switch loadWorkload {
	case "compliant":
		wt = load.WorkloadCompliant
	case "denied":
		wt = load.WorkloadDenied
	case "mixed":
		wt = load.WorkloadMixed
	default:
		return fmt.Errorf("invalid workload type %q: must be compliant, denied, or mixed", loadWorkload)
	}

	var tlsConf *tls.Config
	if loadInsecureTLS {
		tlsConf = &tls.Config{InsecureSkipVerify: true}
	}

	cfg := load.GeneratorConfig{
		TargetURL:       loadTargetURL,
		Workload:        wt,
		Concurrency:     loadConcurrency,
		TotalRequests:   loadRequests,
		TargetQPS:       loadQPS,
		ClientTimeout:   loadTimeout,
		TLSClientConfig: tlsConf,
	}

	if !loadOutputJSON {
		fmt.Printf("Starting KubeGuard Admission Load Test...\n")
		fmt.Printf("Target:      %s\n", loadTargetURL)
		fmt.Printf("Workload:    %s\n", loadWorkload)
		fmt.Printf("Requests:    %d\n", loadRequests)
		fmt.Printf("Concurrency: %d\n\n", loadConcurrency)
	}

	res, _, err := load.RunPhase("sustained", cfg)
	if err != nil {
		return fmt.Errorf("load test failed: %w", err)
	}

	if loadOutputJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(res)
	}

	fmt.Printf("=== Admission Load Test Results ===\n")
	fmt.Printf("Requests:     %d total (%d success, %d errors, %d timeouts)\n",
		res.TotalRequests, res.SuccessCount, res.ErrorCount, res.TimeoutCount)
	fmt.Printf("Decisions:    %d allowed (%d warned), %d denied\n",
		res.AllowedCount, res.WarnedCount, res.DeniedCount)
	fmt.Printf("Duration:     %s\n", res.Duration.Round(time.Millisecond))
	fmt.Printf("Achieved QPS: %.1f req/s\n", res.AchievedQPS)
	fmt.Printf("Latency Min:  %s\n", res.LatencyMin)
	fmt.Printf("Latency p50:  %s\n", res.LatencyP50)
	fmt.Printf("Latency p90:  %s\n", res.LatencyP90)
	fmt.Printf("Latency p99:  %s\n", res.LatencyP99)
	fmt.Printf("Latency Max:  %s\n", res.LatencyMax)

	if res.ErrorCount > 0 || res.TimeoutCount > 0 {
		return fmt.Errorf("load test completed with %d errors and %d timeouts", res.ErrorCount, res.TimeoutCount)
	}

	return nil
}
