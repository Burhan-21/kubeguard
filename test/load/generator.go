package load

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// WorkloadType specifies the manifest character to test.
type WorkloadType string

const (
	WorkloadCompliant WorkloadType = "compliant"
	WorkloadDenied    WorkloadType = "denied"
	WorkloadMixed     WorkloadType = "mixed"
)

// GeneratorConfig configures a load test run.
type GeneratorConfig struct {
	TargetURL       string
	Workload        WorkloadType
	Concurrency     int
	TotalRequests   int
	TargetQPS       int // Rate limiting target (0 = unlimited / maximum concurrency)
	ClientTimeout   time.Duration
	TLSClientConfig *tls.Config
}

// PhaseResult contains aggregated metrics for a load generation phase.
type PhaseResult struct {
	PhaseName     string        `json:"phaseName"`
	TotalRequests int           `json:"totalRequests"`
	SuccessCount  int64         `json:"successCount"`
	AllowedCount  int64         `json:"allowedCount"`
	DeniedCount   int64         `json:"deniedCount"`
	WarnedCount   int64         `json:"warnedCount"`
	ErrorCount    int64         `json:"errorCount"`
	TimeoutCount  int64         `json:"timeoutCount"`
	Duration      time.Duration `json:"duration"`
	AchievedQPS   float64       `json:"achievedQPS"`
	LatencyMin    time.Duration `json:"latencyMin"`
	LatencyP50    time.Duration `json:"latencyP50"`
	LatencyP90    time.Duration `json:"latencyP90"`
	LatencyP99    time.Duration `json:"latencyP99"`
	LatencyMax    time.Duration `json:"latencyMax"`
}

// LoadReport aggregates results across all phases of the test suite.
type LoadReport struct {
	Timestamp      string        `json:"timestamp"`
	TargetURL      string        `json:"targetURL"`
	WorkloadType   WorkloadType  `json:"workloadType"`
	Phases         []PhaseResult `json:"phases"`
	OverallLatency struct {
		P50 time.Duration `json:"p50"`
		P90 time.Duration `json:"p90"`
		P99 time.Duration `json:"p99"`
	} `json:"overallLatency"`
	TotalAllowed int64 `json:"totalAllowed"`
	TotalDenied  int64 `json:"totalDenied"`
	TotalErrors  int64 `json:"totalErrors"`
}

var (
	compliantDeploymentJSON = []byte(`{
		"apiVersion": "apps/v1",
		"kind": "Deployment",
		"metadata": {"name": "load-compliant", "namespace": "default"},
		"spec": {
			"replicas": 2,
			"selector": {"matchLabels": {"app": "load-compliant"}},
			"template": {
				"metadata": {"labels": {"app": "load-compliant"}},
				"spec": {
					"automountServiceAccountToken": false,
					"containers": [{
						"name": "app",
						"image": "registry.example.com/apps/service:v1.0.0@sha256:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
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
						"livenessProbe": {"httpGet": {"path": "/healthz", "port": 8080}},
						"startupProbe": {"httpGet": {"path": "/healthz", "port": 8080}}
					}]
				}
			}
		}
	}`)

	deniedDeploymentJSON = []byte(`{
		"apiVersion": "apps/v1",
		"kind": "Deployment",
		"metadata": {"name": "load-insecure", "namespace": "default"},
		"spec": {
			"replicas": 1,
			"selector": {"matchLabels": {"app": "load-insecure"}},
			"template": {
				"metadata": {"labels": {"app": "load-insecure"}},
				"spec": {
					"hostNetwork": true,
					"containers": [{
						"name": "insecure",
						"image": "nginx:latest",
						"securityContext": {
							"privileged": true
						}
					}]
				}
			}
		}
	}`)
)

// BuildAdmissionReview constructs an AdmissionReview payload for testing.
func BuildAdmissionReview(uid string, isCompliant bool) ([]byte, error) {
	raw := compliantDeploymentJSON
	if !isCompliant {
		raw = deniedDeploymentJSON
	}

	review := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Request: &admissionv1.AdmissionRequest{
			UID: types.UID(uid),
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
			Object: apiruntime.RawExtension{Raw: raw},
		},
	}

	return json.Marshal(review)
}

// RunPhase executes a load test phase against the target webhook.
func RunPhase(phaseName string, cfg GeneratorConfig) (*PhaseResult, []time.Duration, error) {
	timeout := cfg.ClientTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	tr := &http.Transport{
		MaxIdleConns:        cfg.Concurrency * 2,
		MaxIdleConnsPerHost: cfg.Concurrency * 2,
		IdleConnTimeout:     30 * time.Second,
	}
	if cfg.TLSClientConfig != nil {
		tr.TLSClientConfig = cfg.TLSClientConfig
	} else {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // Testing webhook local certificates
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}

	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	var (
		allowedCount atomic.Int64
		deniedCount  atomic.Int64
		warnedCount  atomic.Int64
		errorCount   atomic.Int64
		timeoutCount atomic.Int64
		successCount atomic.Int64
	)

	durations := make([]time.Duration, 0, cfg.TotalRequests)
	var durMu sync.Mutex

	// Work channel
	type reqJob struct {
		index       int
		isCompliant bool
	}
	jobs := make(chan reqJob, cfg.TotalRequests)

	for i := 0; i < cfg.TotalRequests; i++ {
		isComp := true
		switch cfg.Workload {
		case WorkloadDenied:
			isComp = false
		case WorkloadMixed:
			isComp = (i%2 == 0)
		}
		jobs <- reqJob{index: i, isCompliant: isComp}
	}
	close(jobs)

	// Rate limiter ticker if TargetQPS is set
	var ticker *time.Ticker
	if cfg.TargetQPS > 0 {
		interval := time.Second / time.Duration(cfg.TargetQPS)
		ticker = time.NewTicker(interval)
		defer ticker.Stop()
	}

	startTime := time.Now()
	var wg sync.WaitGroup

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobs {
				if ticker != nil {
					<-ticker.C
				}

				uid := fmt.Sprintf("load-%s-%d-%d", phaseName, workerID, job.index)
				payload, err := BuildAdmissionReview(uid, job.isCompliant)
				if err != nil {
					errorCount.Add(1)
					continue
				}

				t0 := time.Now()
				req, err := http.NewRequest(http.MethodPost, cfg.TargetURL, bytes.NewReader(payload))
				if err != nil {
					errorCount.Add(1)
					continue
				}
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				elapsed := time.Since(t0)

				if err != nil {
					if nErr, ok := err.(interface{ Timeout() bool }); ok && nErr.Timeout() {
						timeoutCount.Add(1)
					} else {
						errorCount.Add(1)
					}
					continue
				}

				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					errorCount.Add(1)
					continue
				}

				var reviewResp admissionv1.AdmissionReview
				if err := json.Unmarshal(body, &reviewResp); err != nil || reviewResp.Response == nil {
					errorCount.Add(1)
					continue
				}

				successCount.Add(1)
				if reviewResp.Response.Allowed {
					allowedCount.Add(1)
					if len(reviewResp.Response.Warnings) > 0 {
						warnedCount.Add(1)
					}
				} else {
					deniedCount.Add(1)
				}

				durMu.Lock()
				durations = append(durations, elapsed)
				durMu.Unlock()
			}
		}(w)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	if len(durations) == 0 {
		return &PhaseResult{
			PhaseName:     phaseName,
			TotalRequests: cfg.TotalRequests,
			ErrorCount:    errorCount.Load(),
			TimeoutCount:  timeoutCount.Load(),
			Duration:      totalDuration,
		}, nil, fmt.Errorf("no requests completed successfully")
	}

	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	p50 := sorted[int(math.Round(float64(len(sorted)-1)*0.50))]
	p90 := sorted[int(math.Round(float64(len(sorted)-1)*0.90))]
	p99 := sorted[int(math.Round(float64(len(sorted)-1)*0.99))]

	achievedQPS := float64(successCount.Load()) / totalDuration.Seconds()

	res := &PhaseResult{
		PhaseName:     phaseName,
		TotalRequests: cfg.TotalRequests,
		SuccessCount:  successCount.Load(),
		AllowedCount:  allowedCount.Load(),
		DeniedCount:   deniedCount.Load(),
		WarnedCount:   warnedCount.Load(),
		ErrorCount:    errorCount.Load(),
		TimeoutCount:  timeoutCount.Load(),
		Duration:      totalDuration,
		AchievedQPS:   achievedQPS,
		LatencyMin:    sorted[0],
		LatencyP50:    p50,
		LatencyP90:    p90,
		LatencyP99:    p99,
		LatencyMax:    sorted[len(sorted)-1],
	}

	return res, durations, nil
}
