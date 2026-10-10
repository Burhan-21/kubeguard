package admission

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func generateTestCert(t *testing.T, commonName string, serial int64) (certPEM, keyPEM []byte) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate rsa key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(serial),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:              []string{"localhost", commonName},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("failed to marshal private key: %v", err)
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})

	return certPEM, keyPEM
}

func startTLSTestServer(t *testing.T, reloader *CertReloader) (addr string, cleanup func()) {
	t.Helper()
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		GetCertificate: reloader.GetCertificate,
	})
	if err != nil {
		t.Fatalf("failed to start tls listener: %v", err)
	}

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}
	go server.Serve(ln)

	return ln.Addr().String(), func() {
		server.Close()
		ln.Close()
	}
}

// Test A — Initial certificate
func TestReloader_TestA_InitialCertificate(t *testing.T) {
	certPEM, keyPEM := generateTestCert(t, "kubeguard-init.local", 1001)

	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "tls.crt")
	keyFile := filepath.Join(tmpDir, "tls.key")

	if err := os.WriteFile(certFile, certPEM, 0600); err != nil {
		t.Fatalf("failed to write cert file: %v", err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	reloader, err := NewCertReloader(certFile, keyFile)
	if err != nil {
		t.Fatalf("expected initial load to succeed, got %v", err)
	}

	subject, serial, _, err := reloader.CurrentCertInfo()
	if err != nil {
		t.Fatalf("CurrentCertInfo failed: %v", err)
	}
	if subject != "kubeguard-init.local" {
		t.Errorf("expected subject 'kubeguard-init.local', got %s", subject)
	}
	if serial != "1001" {
		t.Errorf("expected serial '1001', got %s", serial)
	}

	addr, cleanup := startTLSTestServer(t, reloader)
	defer cleanup()

	// Verify TLS handshake served
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         "kubeguard-init.local",
			},
		},
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get("https://" + addr)
	if err != nil {
		t.Fatalf("TLS request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify handshake leaf received by client
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "kubeguard-init.local",
	})
	if err != nil {
		t.Fatalf("tls dial failed: %v", err)
	}
	defer conn.Close()
	peerCerts := conn.ConnectionState().PeerCertificates
	if len(peerCerts) == 0 {
		t.Fatal("no peer certificates returned")
	}
	if peerCerts[0].Subject.CommonName != "kubeguard-init.local" {
		t.Errorf("expected peer cert CN 'kubeguard-init.local', got %s", peerCerts[0].Subject.CommonName)
	}

	// Negative case: invalid initial certificate files must fail immediately
	invalidCertFile := filepath.Join(tmpDir, "invalid.crt")
	os.WriteFile(invalidCertFile, []byte("invalid pem data"), 0600)
	_, err = NewCertReloader(invalidCertFile, keyFile)
	if err == nil {
		t.Error("expected error loading invalid initial certificate, got nil")
	}

	// Non-existent files must fail immediately
	_, err = NewCertReloader(filepath.Join(tmpDir, "nonexistent.crt"), keyFile)
	if err == nil {
		t.Error("expected error for non-existent initial cert file, got nil")
	}
}

// Test B — Certificate rotation
func TestReloader_TestB_CertificateRotation(t *testing.T) {
	certA_PEM, keyA_PEM := generateTestCert(t, "cert-a.local", 2001)
	certB_PEM, keyB_PEM := generateTestCert(t, "cert-b.local", 2002)

	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "tls.crt")
	keyFile := filepath.Join(tmpDir, "tls.key")

	if err := os.WriteFile(certFile, certA_PEM, 0600); err != nil {
		t.Fatalf("failed to write cert A: %v", err)
	}
	if err := os.WriteFile(keyFile, keyA_PEM, 0600); err != nil {
		t.Fatalf("failed to write key A: %v", err)
	}

	reloader, err := NewCertReloader(certFile, keyFile)
	if err != nil {
		t.Fatalf("initial load failed: %v", err)
	}

	addr, cleanup := startTLSTestServer(t, reloader)
	defer cleanup()

	// Verify client receives Certificate A initially
	connA, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "cert-a.local",
	})
	if err != nil {
		t.Fatalf("tls dial failed for cert A: %v", err)
	}
	defer connA.Close()
	if connA.ConnectionState().PeerCertificates[0].Subject.CommonName != "cert-a.local" {
		t.Fatalf("expected cert-a.local, got %s", connA.ConnectionState().PeerCertificates[0].Subject.CommonName)
	}

	// Replace files on disk with Certificate B
	if err := os.WriteFile(certFile, certB_PEM, 0600); err != nil {
		t.Fatalf("failed to update cert file: %v", err)
	}
	if err := os.WriteFile(keyFile, keyB_PEM, 0600); err != nil {
		t.Fatalf("failed to update key file: %v", err)
	}

	// Trigger reload
	if err := reloader.ReloadFromFiles(); err != nil {
		t.Fatalf("reload failed: %v", err)
	}

	// Verify NEW TLS connection receives Certificate B without server restart
	connB, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "cert-b.local",
	})
	if err != nil {
		t.Fatalf("tls dial failed for cert B: %v", err)
	}
	defer connB.Close()
	if connB.ConnectionState().PeerCertificates[0].Subject.CommonName != "cert-b.local" {
		t.Errorf("expected cert-b.local on rotated connection, got %s", connB.ConnectionState().PeerCertificates[0].Subject.CommonName)
	}
	if connB.ConnectionState().PeerCertificates[0].SerialNumber.String() != "2002" {
		t.Errorf("expected serial 2002, got %s", connB.ConnectionState().PeerCertificates[0].SerialNumber.String())
	}
}

// Test C — Invalid rotation fallback
func TestReloader_TestC_InvalidRotationFallback(t *testing.T) {
	certA_PEM, keyA_PEM := generateTestCert(t, "cert-valid.local", 3001)

	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "tls.crt")
	keyFile := filepath.Join(tmpDir, "tls.key")

	os.WriteFile(certFile, certA_PEM, 0600)
	os.WriteFile(keyFile, keyA_PEM, 0600)

	reloader, err := NewCertReloader(certFile, keyFile)
	if err != nil {
		t.Fatalf("initial load failed: %v", err)
	}

	addr, cleanup := startTLSTestServer(t, reloader)
	defer cleanup()

	// Attempt reload with invalid PEM data
	err = reloader.ReloadFromBytes([]byte("corrupted invalid cert data"), keyA_PEM)
	if err == nil {
		t.Fatal("expected error reloading invalid cert, got nil")
	}

	// Attempt reload with mismatched key
	_, keyOther_PEM := generateTestCert(t, "other.local", 3999)
	err = reloader.ReloadFromBytes(certA_PEM, keyOther_PEM)
	if err == nil {
		t.Fatal("expected error reloading mismatched key, got nil")
	}

	// Verify server remains healthy and continues serving Certificate A
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "cert-valid.local",
	})
	if err != nil {
		t.Fatalf("tls dial failed after invalid reload: %v", err)
	}
	defer conn.Close()
	if conn.ConnectionState().PeerCertificates[0].Subject.CommonName != "cert-valid.local" {
		t.Errorf("expected fallback to cert-valid.local, got %s", conn.ConnectionState().PeerCertificates[0].Subject.CommonName)
	}
}

// Test D — Rapid updates
func TestReloader_TestD_RapidUpdates(t *testing.T) {
	reloader, err := NewCertReloaderFromBytes(generateTestCert(t, "rapid-init.local", 4000))
	if err != nil {
		t.Fatalf("initial load failed: %v", err)
	}

	addr, cleanup := startTLSTestServer(t, reloader)
	defer cleanup()

	const iterations = 10
	for i := 1; i <= iterations; i++ {
		cn := fmt.Sprintf("rapid-%d.local", i)
		certPEM, keyPEM := generateTestCert(t, cn, int64(4000+i))
		if err := reloader.ReloadFromBytes(certPEM, keyPEM); err != nil {
			t.Fatalf("rapid update %d failed: %v", i, err)
		}
	}

	// Verify the final certificate is served
	expectedCN := fmt.Sprintf("rapid-%d.local", iterations)
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         expectedCN,
	})
	if err != nil {
		t.Fatalf("tls dial failed: %v", err)
	}
	defer conn.Close()

	actualCN := conn.ConnectionState().PeerCertificates[0].Subject.CommonName
	if actualCN != expectedCN {
		t.Errorf("expected final CN %s, got %s", expectedCN, actualCN)
	}
}

// Test E — Shutdown and Watchers
func TestReloader_TestE_ShutdownAndWatchers(t *testing.T) {
	certA_PEM, keyA_PEM := generateTestCert(t, "filewatch-a.local", 5001)
	certB_PEM, keyB_PEM := generateTestCert(t, "filewatch-b.local", 5002)

	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "tls.crt")
	keyFile := filepath.Join(tmpDir, "tls.key")

	os.WriteFile(certFile, certA_PEM, 0600)
	os.WriteFile(keyFile, keyA_PEM, 0600)

	reloader, err := NewCertReloader(certFile, keyFile)
	if err != nil {
		t.Fatalf("initial load failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Start file watcher with very short interval for test
	go reloader.StartFileWatcher(ctx, 50*time.Millisecond)

	// Update files
	time.Sleep(100 * time.Millisecond)
	os.WriteFile(certFile, certB_PEM, 0600)
	os.WriteFile(keyFile, keyB_PEM, 0600)

	// Wait for file watcher to pick up change
	var rotated bool
	for attempt := 0; attempt < 20; attempt++ {
		time.Sleep(50 * time.Millisecond)
		subject, _, _, _ := reloader.CurrentCertInfo()
		if subject == "filewatch-b.local" {
			rotated = true
			break
		}
	}
	if !rotated {
		t.Error("file watcher did not rotate certificate within expected duration")
	}

	// Cancel context and verify shutdown
	cancel()
	time.Sleep(100 * time.Millisecond)
}

// Test Kubernetes Secret Watcher with Fake Client
func TestReloader_KubernetesSecretWatcher(t *testing.T) {
	certA_PEM, keyA_PEM := generateTestCert(t, "k8s-cert-a.local", 6001)
	certB_PEM, keyB_PEM := generateTestCert(t, "k8s-cert-b.local", 6002)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kubeguard-tls",
			Namespace: "kubeguard-system",
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": certA_PEM,
			"tls.key": keyA_PEM,
		},
	}

	fakeClient := fake.NewSimpleClientset(secret)

	reloader, err := NewCertReloaderFromBytes(certA_PEM, keyA_PEM)
	if err != nil {
		t.Fatalf("initial load failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go reloader.StartSecretWatcher(ctx, fakeClient, "kubeguard-system", "kubeguard-tls")

	time.Sleep(100 * time.Millisecond)

	// Update Secret in fake client
	secret.Data["tls.crt"] = certB_PEM
	secret.Data["tls.key"] = keyB_PEM
	_, err = fakeClient.CoreV1().Secrets("kubeguard-system").Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("failed to update secret: %v", err)
	}

	// Wait for secret watcher event
	var rotated bool
	for attempt := 0; attempt < 20; attempt++ {
		time.Sleep(50 * time.Millisecond)
		subject, _, _, _ := reloader.CurrentCertInfo()
		if subject == "k8s-cert-b.local" {
			rotated = true
			break
		}
	}
	if !rotated {
		t.Error("Kubernetes secret watcher did not rotate certificate within expected duration")
	}
}

// Test F — Concurrency & Race Safety
func TestReloader_TestF_RaceSafety(t *testing.T) {
	certA_PEM, keyA_PEM := generateTestCert(t, "race-init.local", 7000)
	reloader, err := NewCertReloaderFromBytes(certA_PEM, keyA_PEM)
	if err != nil {
		t.Fatalf("initial load failed: %v", err)
	}

	addr, cleanup := startTLSTestServer(t, reloader)
	defer cleanup()

	const clientGoroutines = 8
	const reloaderGoroutines = 4
	const iterations = 25

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Readers: perform continuous TLS handshakes
	for g := 0; g < clientGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
				Timeout: 2 * time.Second,
			}
			for i := 0; i < iterations; i++ {
				select {
				case <-ctx.Done():
					return
				default:
					resp, err := client.Get("https://" + addr)
					if err == nil {
						resp.Body.Close()
					}
					time.Sleep(5 * time.Millisecond)
				}
			}
		}()
	}

	// Writers: concurrently rotate valid and invalid certificates
	for g := 0; g < reloaderGoroutines; g++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				select {
				case <-ctx.Done():
					return
				default:
					if i%3 == 0 {
						// Attempt invalid reload (should safely be rejected)
						_ = reloader.ReloadFromBytes([]byte("invalid pem data"), []byte("invalid key"))
					} else {
						// Valid rotation
						cn := fmt.Sprintf("race-w%d-i%d.local", writerID, i)
						certPEM, keyPEM := generateTestCert(t, cn, int64(8000+writerID*100+i))
						_ = reloader.ReloadFromBytes(certPEM, keyPEM)
					}
					time.Sleep(10 * time.Millisecond)
				}
			}
		}(g)
	}

	wg.Wait()

	// Ensure server is still functioning after race test
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Get("https://" + addr)
	if err != nil {
		t.Fatalf("server unhealthy after concurrent rotation test: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
