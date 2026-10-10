package load

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Burhan-21/kubeguard/internal/admission"
	"github.com/Burhan-21/kubeguard/internal/policy"
	"github.com/Burhan-21/kubeguard/internal/rules/reliability"
	"github.com/Burhan-21/kubeguard/internal/rules/security"
)

// setupMockAdmissionServer starts a live admission server on a random local port.
func setupMockAdmissionServer(t *testing.T) (*admission.Server, string, func()) {
	var rules []policy.Rule
	for _, r := range security.AllSecurityRules() {
		rules = append(rules, r)
	}
	for _, r := range reliability.AllReliabilityRules() {
		rules = append(rules, r)
	}
	engine := policy.NewEngine(rules)
	handler := &admission.Handler{Engine: engine, Mode: "enforce"}

	// Find free port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()

	// Generate temporary TLS certs
	certFile, keyFile, cleanupCerts := generateTempTLSCert(t)

	server := &admission.Server{
		Handler:  handler,
		Port:     port,
		CertFile: certFile,
		KeyFile:  keyFile,
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)

	go func() {
		errCh <- server.StartWithContext(ctx)
	}()

	url := fmt.Sprintf("https://127.0.0.1:%d/validate", port)

	// Wait for server readiness
	readyURL := fmt.Sprintf("https://127.0.0.1:%d/readyz", port)
	client := &http.Client{
		Timeout: 200 * time.Millisecond,
		Transport: &http.Transport{
			TLSClientConfig: nil, // server uses self-signed
		},
	}
	// For testing, accept self-signed
	http.DefaultTransport.(*http.Transport).TLSClientConfig = nil

	ready := false
	for i := 0; i < 30; i++ {
		req, _ := http.NewRequest(http.MethodGet, readyURL, nil)
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			ready = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if !ready {
		cancel()
		cleanupCerts()
		t.Fatalf("admission test server failed to become ready")
	}

	cleanup := func() {
		cancel()
		_ = server.Shutdown(context.Background())
		cleanupCerts()
	}

	return server, url, cleanup
}

func generateTempTLSCert(t *testing.T) (string, string, func()) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa generate key failed: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"KubeGuard Test"},
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(24 * time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("create certificate failed: %v", err)
	}

	certFile, err := os.CreateTemp("", "cert-*.pem")
	if err != nil {
		t.Fatalf("temp cert file: %v", err)
	}
	_ = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certFile.Close()

	keyFile, err := os.CreateTemp("", "key-*.pem")
	if err != nil {
		t.Fatalf("temp key file: %v", err)
	}
	privBytes, _ := x509.MarshalPKCS8PrivateKey(priv)
	_ = pem.Encode(keyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	keyFile.Close()

	cleanup := func() {
		_ = os.Remove(certFile.Name())
		_ = os.Remove(keyFile.Name())
	}

	return certFile.Name(), keyFile.Name(), cleanup
}

func TestLoadGeneratorPhases(t *testing.T) {
	_, url, cleanup := setupMockAdmissionServer(t)
	defer cleanup()

	// 1. Warm-up Phase
	warmupCfg := GeneratorConfig{
		TargetURL:     url,
		Workload:      WorkloadCompliant,
		Concurrency:   2,
		TotalRequests: 50,
		TargetQPS:     0,
		ClientTimeout: 2 * time.Second,
	}
	warmupRes, _, err := RunPhase("warm-up", warmupCfg)
	if err != nil {
		t.Fatalf("warm-up failed: %v", err)
	}
	if warmupRes.SuccessCount != 50 || warmupRes.ErrorCount != 0 {
		t.Errorf("warm-up unexpected counts: %+v", warmupRes)
	}

	// 2. Sustained Phase (Mixed Workloads)
	sustainedCfg := GeneratorConfig{
		TargetURL:     url,
		Workload:      WorkloadMixed,
		Concurrency:   5,
		TotalRequests: 100,
		TargetQPS:     0,
		ClientTimeout: 2 * time.Second,
	}
	sustainedRes, _, err := RunPhase("sustained", sustainedCfg)
	if err != nil {
		t.Fatalf("sustained phase failed: %v", err)
	}
	if sustainedRes.AllowedCount == 0 || sustainedRes.DeniedCount == 0 {
		t.Errorf("expected mixed decisions, got allowed=%d, denied=%d", sustainedRes.AllowedCount, sustainedRes.DeniedCount)
	}

	// 3. Burst Phase
	burstCfg := GeneratorConfig{
		TargetURL:     url,
		Workload:      WorkloadCompliant,
		Concurrency:   10,
		TotalRequests: 100,
		TargetQPS:     0,
		ClientTimeout: 2 * time.Second,
	}
	burstRes, _, err := RunPhase("burst", burstCfg)
	if err != nil {
		t.Fatalf("burst phase failed: %v", err)
	}
	if burstRes.ErrorCount != 0 {
		t.Errorf("burst phase had errors: %d", burstRes.ErrorCount)
	}

	// 4. Recovery Phase
	recoveryCfg := GeneratorConfig{
		TargetURL:     url,
		Workload:      WorkloadCompliant,
		Concurrency:   2,
		TotalRequests: 20,
		TargetQPS:     0,
		ClientTimeout: 2 * time.Second,
	}
	recoveryRes, _, err := RunPhase("recovery", recoveryCfg)
	if err != nil {
		t.Fatalf("recovery phase failed: %v", err)
	}
	if recoveryRes.ErrorCount != 0 {
		t.Errorf("recovery phase had errors: %d", recoveryRes.ErrorCount)
	}
}
