package admission

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics records operational metrics for the admission webhook.
// All counters and gauges use atomic operations for lock-free concurrency safety.
type Metrics struct {
	// Request counters
	requestsTotal      atomic.Uint64
	requestsAllowed    atomic.Uint64
	requestsDenied     atomic.Uint64
	requestsWarned     atomic.Uint64
	requestErrorsTotal atomic.Uint64

	// Latency tracking (cumulative seconds and count)
	requestDurationSeconds atomic.Uint64 // stored as microseconds for atomic operations

	// Prometheus standard histogram buckets for admission request duration (seconds)
	// Buckets: 0.001, 0.005, 0.010, 0.025, 0.050, 0.100, 0.250, 0.500, 1.0, +Inf
	durationBucketCounts [9]atomic.Uint64
	durationHistCount    atomic.Uint64
	durationHistSumNanos atomic.Uint64 // nanoseconds for precision

	// Policy evaluation errors
	policyErrorsTotal atomic.Uint64

	// TLS certificate reload counters
	certReloadSuccessTotal atomic.Uint64
	certReloadFailureTotal atomic.Uint64

	// Cert expiry Unix timestamp in seconds
	certExpiryTimestamp atomic.Int64

	// mu guards metric scrape rendering
	mu sync.RWMutex
}

// DurationHistogramBuckets defines the upper bounds (in seconds) for admission duration histogram.
var DurationHistogramBuckets = []float64{0.001, 0.005, 0.010, 0.025, 0.050, 0.100, 0.250, 0.500, 1.0}

var defaultMetrics = NewMetrics()

// DefaultMetrics returns the global metrics collector.
func DefaultMetrics() *Metrics {
	return defaultMetrics
}

// NewMetrics initializes an empty Metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{}
}

// IncRequestsTotal increments total admission requests.
func (m *Metrics) IncRequestsTotal() {
	m.requestsTotal.Add(1)
}

// IncRequestsAllowed increments allowed admission decisions.
func (m *Metrics) IncRequestsAllowed() {
	m.requestsAllowed.Add(1)
}

// IncRequestsDenied increments denied admission decisions.
func (m *Metrics) IncRequestsDenied() {
	m.requestsDenied.Add(1)
}

// IncRequestsWarned increments warned admission decisions.
func (m *Metrics) IncRequestsWarned() {
	m.requestsWarned.Add(1)
}

// IncRequestErrorsTotal increments malformed or bad requests.
func (m *Metrics) IncRequestErrorsTotal() {
	m.requestErrorsTotal.Add(1)
}

// ObserveRequestDuration records the request duration across both cumulative counter and standard histogram buckets.
func (m *Metrics) ObserveRequestDuration(d time.Duration) {
	secs := d.Seconds()
	micros := uint64(d.Microseconds())
	nanos := uint64(d.Nanoseconds())

	// 1. Maintain backward-compatible cumulative duration counter
	m.requestDurationSeconds.Add(micros)

	// 2. Histogram count and sum
	m.durationHistCount.Add(1)
	m.durationHistSumNanos.Add(nanos)

	// 3. Increment matching histogram buckets
	for i, upper := range DurationHistogramBuckets {
		if secs <= upper {
			m.durationBucketCounts[i].Add(1)
		}
	}
}

// IncPolicyErrorsTotal increments policy engine evaluation errors.
func (m *Metrics) IncPolicyErrorsTotal() {
	m.policyErrorsTotal.Add(1)
}

// IncCertReloadSuccess increments successful TLS certificate reloads.
func (m *Metrics) IncCertReloadSuccess() {
	m.certReloadSuccessTotal.Add(1)
}

// IncCertReloadFailure increments failed TLS certificate reloads.
func (m *Metrics) IncCertReloadFailure() {
	m.certReloadFailureTotal.Add(1)
}

// SetCertExpiry records the active certificate's expiry as a Unix timestamp.
func (m *Metrics) SetCertExpiry(t time.Time) {
	m.certExpiryTimestamp.Store(t.Unix())
}

// WritePrometheus writes metrics in Prometheus standard text exposition format.
// Strictly avoids high-cardinality labels or sensitive request information.
func (m *Metrics) WritePrometheus(w io.Writer) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	reqTotal := m.requestsTotal.Load()
	reqAllowed := m.requestsAllowed.Load()
	reqDenied := m.requestsDenied.Load()
	reqWarned := m.requestsWarned.Load()
	reqErrors := m.requestErrorsTotal.Load()
	durMicros := m.requestDurationSeconds.Load()
	policyErr := m.policyErrorsTotal.Load()
	reloadSucc := m.certReloadSuccessTotal.Load()
	reloadFail := m.certReloadFailureTotal.Load()
	certExp := m.certExpiryTimestamp.Load()

	// Snapshot histogram counters
	histCount := m.durationHistCount.Load()
	histSumSecs := float64(m.durationHistSumNanos.Load()) / 1e9
	var bucketCounts [9]uint64
	for i := range bucketCounts {
		bucketCounts[i] = m.durationBucketCounts[i].Load()
	}

	// Admission requests total by decision
	fmt.Fprintln(w, "# HELP kubeguard_admission_requests_total Total number of admission requests processed by decision.")
	fmt.Fprintln(w, "# TYPE kubeguard_admission_requests_total counter")
	fmt.Fprintf(w, "kubeguard_admission_requests_total{decision=\"allowed\"} %d\n", reqAllowed)
	fmt.Fprintf(w, "kubeguard_admission_requests_total{decision=\"denied\"} %d\n", reqDenied)
	fmt.Fprintf(w, "kubeguard_admission_requests_total{decision=\"warned\"} %d\n", reqWarned)
	fmt.Fprintf(w, "kubeguard_admission_requests_total{decision=\"all\"} %d\n", reqTotal)

	// Admission request errors
	fmt.Fprintln(w, "# HELP kubeguard_admission_request_errors_total Total number of malformed or invalid admission requests.")
	fmt.Fprintln(w, "# TYPE kubeguard_admission_request_errors_total counter")
	fmt.Fprintf(w, "kubeguard_admission_request_errors_total %d\n", reqErrors)

	// Admission request duration (backward-compatible counter)
	durSec := float64(durMicros) / 1000000.0
	fmt.Fprintln(w, "# HELP kubeguard_admission_request_duration_seconds_total Total duration of admission request processing in seconds.")
	fmt.Fprintln(w, "# TYPE kubeguard_admission_request_duration_seconds_total counter")
	fmt.Fprintf(w, "kubeguard_admission_request_duration_seconds_total %.6f\n", durSec)

	// Prometheus standard histogram
	fmt.Fprintln(w, "# HELP kubeguard_admission_request_duration_seconds Admission request duration histogram in seconds.")
	fmt.Fprintln(w, "# TYPE kubeguard_admission_request_duration_seconds histogram")
	for i, upper := range DurationHistogramBuckets {
		fmt.Fprintf(w, "kubeguard_admission_request_duration_seconds_bucket{le=\"%s\"} %d\n", formatBucketBound(upper), bucketCounts[i])
	}
	fmt.Fprintf(w, "kubeguard_admission_request_duration_seconds_bucket{le=\"+Inf\"} %d\n", histCount)
	fmt.Fprintf(w, "kubeguard_admission_request_duration_seconds_sum %.6f\n", histSumSecs)
	fmt.Fprintf(w, "kubeguard_admission_request_duration_seconds_count %d\n", histCount)

	// Policy evaluation errors
	fmt.Fprintln(w, "# HELP kubeguard_policy_errors_total Total number of errors encountered during policy evaluation.")
	fmt.Fprintln(w, "# TYPE kubeguard_policy_errors_total counter")
	fmt.Fprintf(w, "kubeguard_policy_errors_total %d\n", policyErr)

	// TLS certificate reload operations
	fmt.Fprintln(w, "# HELP kubeguard_tls_certificate_reloads_total Total number of TLS certificate reload attempts by result.")
	fmt.Fprintln(w, "# TYPE kubeguard_tls_certificate_reloads_total counter")
	fmt.Fprintf(w, "kubeguard_tls_certificate_reloads_total{result=\"success\"} %d\n", reloadSucc)
	fmt.Fprintf(w, "kubeguard_tls_certificate_reloads_total{result=\"failure\"} %d\n", reloadFail)

	// TLS certificate expiry timestamp
	fmt.Fprintln(w, "# HELP kubeguard_tls_certificate_expiry_timestamp_seconds Expiry timestamp of the currently active TLS certificate in seconds since Unix epoch.")
	fmt.Fprintln(w, "# TYPE kubeguard_tls_certificate_expiry_timestamp_seconds gauge")
	fmt.Fprintf(w, "kubeguard_tls_certificate_expiry_timestamp_seconds %d\n", certExp)

	// Process health gauge
	fmt.Fprintln(w, "# HELP kubeguard_webhook_healthy Readiness and health status of the webhook process.")
	fmt.Fprintln(w, "# TYPE kubeguard_webhook_healthy gauge")
	fmt.Fprintln(w, "kubeguard_webhook_healthy 1")
}

func formatBucketBound(f float64) string {
	return fmt.Sprintf("%.3f", f)
}
