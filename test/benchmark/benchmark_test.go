package benchmark

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"runtime"
	"sort"
	"testing"
	"time"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"

	"github.com/Burhan-21/kubeguard/internal/admission"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
	"github.com/Burhan-21/kubeguard/internal/parser"
	"github.com/Burhan-21/kubeguard/internal/policy"
	"github.com/Burhan-21/kubeguard/internal/reporter"
	"github.com/Burhan-21/kubeguard/internal/rules/reliability"
	"github.com/Burhan-21/kubeguard/internal/rules/security"
)

// generateManifests creates n representative Kubernetes Deployment manifests in multi-doc YAML format.
func generateManifests(count int) []byte {
	var buf bytes.Buffer
	for i := 1; i <= count; i++ {
		buf.WriteString(fmt.Sprintf(`---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: workload-%d
  namespace: default
  labels:
    app: workload-%d
spec:
  replicas: 3
  selector:
    matchLabels:
      app: workload-%d
  template:
    metadata:
      labels:
        app: workload-%d
    spec:
      automountServiceAccountToken: false
      containers:
      - name: app
        image: registry.example.com/apps/service-%d:v1.2.0@sha256:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
        securityContext:
          privileged: false
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          runAsNonRoot: true
          runAsUser: 10001
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 200m
            memory: 256Mi
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
        startupProbe:
          httpGet:
            path: /healthz
            port: 8080
`, i, i, i, i, i))
	}
	return buf.Bytes()
}

func getEngine() *policy.Engine {
	var rules []policy.Rule
	for _, r := range security.AllSecurityRules() {
		rules = append(rules, r)
	}
	for _, r := range reliability.AllReliabilityRules() {
		rules = append(rules, r)
	}
	return policy.NewEngine(rules)
}

type Stats struct {
	Count       int
	Durations   []time.Duration
	Min         time.Duration
	P50         time.Duration
	P95         time.Duration
	P99         time.Duration
	Max         time.Duration
	Throughput  float64 // resources per second
	AllocBytes  uint64
	TotalAllocs uint64
}

func calculateStats(durations []time.Duration, resourceCount int, allocBytes, allocCount uint64) Stats {
	if len(durations) == 0 {
		return Stats{}
	}
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	p50 := sorted[int(math.Round(float64(len(sorted)-1)*0.50))]
	p95 := sorted[int(math.Round(float64(len(sorted)-1)*0.95))]
	p99 := sorted[int(math.Round(float64(len(sorted)-1)*0.99))]

	medianSec := p50.Seconds()
	var throughput float64
	if medianSec > 0 {
		throughput = float64(resourceCount) / medianSec
	}

	return Stats{
		Count:       len(durations),
		Durations:   sorted,
		Min:         sorted[0],
		P50:         p50,
		P95:         p95,
		P99:         p99,
		Max:         sorted[len(sorted)-1],
		Throughput:  throughput,
		AllocBytes:  allocBytes / uint64(len(durations)),
		TotalAllocs: allocCount / uint64(len(durations)),
	}
}

// runScanPipeline parses, normalizes, evaluates, and serializes JSON reports.
func runScanPipeline(yamlData []byte, engine *policy.Engine) (int, error) {
	unstructuredList, err := parser.ParseReader(bytes.NewReader(yamlData))
	if err != nil {
		return 0, err
	}

	var normalizedList []*normalizer.NormalizedResource
	for _, obj := range unstructuredList {
		norm, normErr := normalizer.Normalize(obj)
		if normErr != nil {
			continue
		}
		normalizedList = append(normalizedList, norm)
	}

	scanResult := engine.EvaluateAll(normalizedList)
	if err := reporter.ReportJSON(scanResult, io.Discard); err != nil {
		return 0, err
	}

	return len(normalizedList), nil
}

func TestPerformanceSuite(t *testing.T) {
	engine := getEngine()

	testCases := []struct {
		name          string
		resourceCount int
		iterations    int
		targetMax     time.Duration
	}{
		{name: "CLI Scan 10 Resources", resourceCount: 10, iterations: 20, targetMax: 100 * time.Millisecond},
		{name: "CLI Scan 100 Resources", resourceCount: 100, iterations: 20, targetMax: 1 * time.Second},
		{name: "CLI Scan 1000 Resources", resourceCount: 1000, iterations: 10, targetMax: 10 * time.Second},
	}

	fmt.Println("\n=========================================================================================")
	fmt.Println("                             KUBEGUARD PERFORMANCE BENCHMARK REPORT                      ")
	fmt.Println("=========================================================================================")
	fmt.Printf("%-24s | %6s | %10s | %10s | %10s | %10s | %12s | %8s\n",
		"Benchmark", "Count", "Min", "p50 (Med)", "p95", "Max", "Throughput", "Target")
	fmt.Println("-----------------------------------------------------------------------------------------")

	for _, tc := range testCases {
		yamlData := generateManifests(tc.resourceCount)

		// Warmup
		_, _ = runScanPipeline(yamlData, engine)

		var mStart, mEnd runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&mStart)

		var durations []time.Duration
		for i := 0; i < tc.iterations; i++ {
			t0 := time.Now()
			n, err := runScanPipeline(yamlData, engine)
			dur := time.Since(t0)
			if err != nil {
				t.Fatalf("%s failed: %v", tc.name, err)
			}
			if n != tc.resourceCount {
				t.Fatalf("%s processed %d resources, expected %d", tc.name, n, tc.resourceCount)
			}
			durations = append(durations, dur)
		}

		runtime.ReadMemStats(&mEnd)
		allocBytes := mEnd.TotalAlloc - mStart.TotalAlloc
		allocCount := mEnd.Mallocs - mStart.Mallocs

		stats := calculateStats(durations, tc.resourceCount, allocBytes, allocCount)

		targetStatus := "PASS"
		if stats.P50 > tc.targetMax {
			targetStatus = "FAIL"
		}

		fmt.Printf("%-24s | %6d | %10s | %10s | %10s | %10s | %9.1f r/s | %s (<%s)\n",
			tc.name, tc.resourceCount, stats.Min, stats.P50, stats.P95, stats.Max, stats.Throughput, targetStatus, tc.targetMax)

		if stats.P50 > tc.targetMax {
			t.Errorf("%s exceeded documented threshold: p50 %s > %s", tc.name, stats.P50, tc.targetMax)
		}
	}

	// Benchmark Admission Handler Evaluation (Isolated policy evaluation latency)
	admissionDeploymentJSON := []byte(`{
		"apiVersion": "apps/v1",
		"kind": "Deployment",
		"metadata": {"name": "admission-bench", "namespace": "default"},
		"spec": {
			"replicas": 2,
			"selector": {"matchLabels": {"app": "bench"}},
			"template": {
				"metadata": {"labels": {"app": "bench"}},
				"spec": {
					"automountServiceAccountToken": false,
					"containers": [{
						"name": "app",
						"image": "app:v1.0.0@sha256:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
						"securityContext": {
							"privileged": false,
							"allowPrivilegeEscalation": false,
							"readOnlyRootFilesystem": true,
							"runAsNonRoot": true,
							"runAsUser": 10001
						},
						"resources": {
							"requests": {"cpu": "100m", "memory": "128Mi"},
							"limits": {"cpu": "200m", "memory": "256Mi"}
						},
						"readinessProbe": {"httpGet": {"path": "/ready", "port": 8080}},
						"livenessProbe": {"httpGet": {"path": "/healthz", "port": 8080}}
					}]
				}
			}
		}
	}`)

	handler := &admission.Handler{
		Engine: engine,
		Mode:   "enforce",
	}

	review := &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Request: &admissionv1.AdmissionRequest{
			UID: "bench-uid-1234",
			Kind: metav1.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			},
			Resource: metav1.GroupVersionResource{
				Group:    "apps",
				Version:  "v1",
				Resource: "deployments",
			},
			Object: apiruntime.RawExtension{Raw: admissionDeploymentJSON},
		},
	}

	// Warmup
	_ = handler.Handle(review)

	admissionIters := 100
	var admissionDurations []time.Duration
	for i := 0; i < admissionIters; i++ {
		t0 := time.Now()
		resp := handler.Handle(review)
		dur := time.Since(t0)
		if !resp.Allowed {
			t.Fatalf("admission handler unexpectedly rejected compliant deployment")
		}
		admissionDurations = append(admissionDurations, dur)
	}

	admStats := calculateStats(admissionDurations, 1, 0, 0)
	fmt.Printf("%-24s | %6d | %10s | %10s | %10s | %10s | %9.1f r/s | MEASURED\n",
		"Admission Handler Eval", 1, admStats.Min, admStats.P50, admStats.P95, admStats.Max, admStats.Throughput)
	fmt.Println("=========================================================================================")
}

// Go benchmark implementations
func BenchmarkScan10(b *testing.B) {
	engine := getEngine()
	data := generateManifests(10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = runScanPipeline(data, engine)
	}
}

func BenchmarkScan100(b *testing.B) {
	engine := getEngine()
	data := generateManifests(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = runScanPipeline(data, engine)
	}
}

func BenchmarkAdmissionHandler(b *testing.B) {
	engine := getEngine()
	handler := &admission.Handler{Engine: engine, Mode: "enforce"}
	review := &admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:    "bench-uid",
			Object: apiruntime.RawExtension{Raw: []byte(`{"apiVersion":"v1","kind":"Pod","metadata":{"name":"test"},"spec":{"containers":[{"name":"c","image":"img@sha256:1111111111111111111111111111111111111111111111111111111111111111"}]}}`)},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.Handle(review)
	}
}
