package admission

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestMetrics_PrometheusExposition(t *testing.T) {
	m := NewMetrics()

	// Initial output
	var buf bytes.Buffer
	m.WritePrometheus(&buf)
	out := buf.String()

	if !strings.Contains(out, "kubeguard_admission_requests_total{decision=\"all\"} 0") {
		t.Errorf("expected 0 total requests, got:\n%s", out)
	}
	if !strings.Contains(out, "kubeguard_webhook_healthy 1") {
		t.Errorf("expected webhook healthy gauge, got:\n%s", out)
	}

	// Record actions
	m.IncRequestsTotal()
	m.IncRequestsAllowed()
	m.IncRequestsDenied()
	m.IncRequestsWarned()
	m.IncRequestErrorsTotal()
	m.IncPolicyErrorsTotal()
	m.ObserveRequestDuration(5 * time.Millisecond)
	m.IncCertReloadSuccess()
	m.IncCertReloadFailure()
	fixedTime := time.Unix(1791500000, 0)
	m.SetCertExpiry(fixedTime)

	buf.Reset()
	m.WritePrometheus(&buf)
	out = buf.String()

	expectedSubstrings := []string{
		"kubeguard_admission_requests_total{decision=\"allowed\"} 1",
		"kubeguard_admission_requests_total{decision=\"denied\"} 1",
		"kubeguard_admission_requests_total{decision=\"warned\"} 1",
		"kubeguard_admission_requests_total{decision=\"all\"} 1",
		"kubeguard_admission_request_errors_total 1",
		"kubeguard_policy_errors_total 1",
		"kubeguard_admission_request_duration_seconds_bucket{le=\"0.001\"} 0",
		"kubeguard_admission_request_duration_seconds_bucket{le=\"0.005\"} 1",
		"kubeguard_admission_request_duration_seconds_bucket{le=\"0.010\"} 1",
		"kubeguard_admission_request_duration_seconds_bucket{le=\"+Inf\"} 1",
		"kubeguard_admission_request_duration_seconds_count 1",
		"kubeguard_tls_certificate_reloads_total{result=\"success\"} 1",
		"kubeguard_tls_certificate_reloads_total{result=\"failure\"} 1",
		"kubeguard_tls_certificate_expiry_timestamp_seconds 1791500000",
		"kubeguard_webhook_healthy 1",
	}

	for _, exp := range expectedSubstrings {
		if !strings.Contains(out, exp) {
			t.Errorf("expected output to contain %q, but got:\n%s", exp, out)
		}
	}
}

func TestMetrics_Concurrency(t *testing.T) {
	m := NewMetrics()
	done := make(chan bool)

	// Spin up 10 goroutines recording metrics concurrently
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				m.IncRequestsTotal()
				m.IncRequestsAllowed()
				m.ObserveRequestDuration(100 * time.Microsecond)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	var buf bytes.Buffer
	m.WritePrometheus(&buf)
	out := buf.String()

	if !strings.Contains(out, "kubeguard_admission_requests_total{decision=\"all\"} 1000") {
		t.Errorf("expected 1000 total requests after concurrent increments, got:\n%s", out)
	}
	if !strings.Contains(out, "kubeguard_admission_requests_total{decision=\"allowed\"} 1000") {
		t.Errorf("expected 1000 allowed requests after concurrent increments, got:\n%s", out)
	}
}
