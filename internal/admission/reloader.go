package admission

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// CertReloader provides atomic, dynamic TLS certificate rotation with in-memory reload.
// It retains the last known-good certificate if an invalid update is encountered.
type CertReloader struct {
	currentCert atomic.Pointer[tls.Certificate]
	certHash    atomic.Pointer[string]
	mu          sync.Mutex // serializes reload operations
	certFile    string
	keyFile     string
}

// NewCertReloader creates a reloader initialized with certificate and key files from disk.
// Startup fails immediately if initial files are invalid or unreadable.
func NewCertReloader(certFile, keyFile string) (*CertReloader, error) {
	if certFile == "" || keyFile == "" {
		return nil, errors.New("both certFile and keyFile must be specified")
	}

	r := &CertReloader{
		certFile: certFile,
		keyFile:  keyFile,
	}

	if err := r.ReloadFromFiles(); err != nil {
		return nil, fmt.Errorf("initial certificate load failed: %w", err)
	}

	return r, nil
}

// NewCertReloaderFromBytes creates a reloader initialized with in-memory PEM bytes.
// Fails immediately if initial PEM bytes are invalid.
func NewCertReloaderFromBytes(certPEM, keyPEM []byte) (*CertReloader, error) {
	r := &CertReloader{}
	if err := r.ReloadFromBytes(certPEM, keyPEM); err != nil {
		return nil, fmt.Errorf("initial certificate load from bytes failed: %w", err)
	}
	return r, nil
}

// GetCertificate dynamically returns the current active TLS certificate for client handshakes.
// Concurrency-safe and lock-free on the read path.
func (r *CertReloader) GetCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	cert := r.currentCert.Load()
	if cert == nil {
		return nil, errors.New("no active TLS certificate available")
	}
	return cert, nil
}

// ReloadFromFiles reads certificate and key files from disk, validates them, and atomically swaps.
// If reload fails, the last known-good certificate is retained.
func (r *CertReloader) ReloadFromFiles() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	certPEM, err := os.ReadFile(r.certFile)
	if err != nil {
		return fmt.Errorf("failed to read cert file %s: %w", r.certFile, err)
	}

	keyPEM, err := os.ReadFile(r.keyFile)
	if err != nil {
		return fmt.Errorf("failed to read key file %s: %w", r.keyFile, err)
	}

	return r.applyLocked(certPEM, keyPEM, fmt.Sprintf("files (%s, %s)", r.certFile, r.keyFile))
}

// ReloadFromBytes validates in-memory PEM bytes and atomically updates the active certificate.
// If reload fails, the last known-good certificate is retained.
func (r *CertReloader) ReloadFromBytes(certPEM, keyPEM []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.applyLocked(certPEM, keyPEM, "in-memory bytes")
}

// applyLocked validates the certificate pair and replaces the active certificate atomically.
// Must be called with r.mu held.
func (r *CertReloader) applyLocked(certPEM, keyPEM []byte, source string) error {
	if len(certPEM) == 0 || len(keyPEM) == 0 {
		return errors.New("certificate or private key PEM data is empty")
	}

	// Compute checksum to avoid redundant parsing/swapping
	hasher := sha256.New()
	hasher.Write(certPEM)
	hasher.Write([]byte{0})
	hasher.Write(keyPEM)
	newHash := hex.EncodeToString(hasher.Sum(nil))

	oldHashPtr := r.certHash.Load()
	if oldHashPtr != nil && *oldHashPtr == newHash {
		// Content unchanged, no action needed
		return nil
	}

	// Parse TLS key pair
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		DefaultMetrics().IncCertReloadFailure()
		return fmt.Errorf("invalid x509 key pair: %w", err)
	}

	if len(tlsCert.Certificate) == 0 {
		DefaultMetrics().IncCertReloadFailure()
		return errors.New("parsed TLS certificate contains zero certificates in chain")
	}

	// Parse and validate leaf certificate
	leaf, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		DefaultMetrics().IncCertReloadFailure()
		return fmt.Errorf("failed to parse x509 leaf certificate: %w", err)
	}
	tlsCert.Leaf = leaf

	// Atomically swap certificate pointer and hash
	r.currentCert.Store(&tlsCert)
	r.certHash.Store(&newHash)

	DefaultMetrics().IncCertReloadSuccess()
	DefaultMetrics().SetCertExpiry(leaf.NotAfter)

	log.Printf("[TLS] Certificate successfully loaded from %s (Subject: %s, Serial: %s, NotAfter: %s)",
		source, leaf.Subject.CommonName, leaf.SerialNumber.String(), leaf.NotAfter.Format(time.RFC3339))

	return nil
}

// CurrentCertInfo returns non-sensitive metadata about the currently active certificate.
func (r *CertReloader) CurrentCertInfo() (subject string, serial string, notAfter time.Time, err error) {
	cert := r.currentCert.Load()
	if cert == nil || cert.Leaf == nil {
		return "", "", time.Time{}, errors.New("no certificate loaded")
	}
	return cert.Leaf.Subject.CommonName, cert.Leaf.SerialNumber.String(), cert.Leaf.NotAfter, nil
}

// StartFileWatcher periodically checks certificate files for changes and reloads them.
// Runs until ctx is cancelled.
func (r *CertReloader) StartFileWatcher(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 2 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[TLS] File watcher stopped")
			return
		case <-ticker.C:
			if err := r.ReloadFromFiles(); err != nil {
				log.Printf("[TLS] Warning: Failed to reload certificate from files: %v; retaining last known-good certificate", err)
			}
		}
	}
}

// StartSecretWatcher watches a specific Kubernetes Secret and reloads certificates on update.
// Runs until ctx is cancelled.
func (r *CertReloader) StartSecretWatcher(ctx context.Context, kubeClient kubernetes.Interface, namespace, secretName string) {
	if kubeClient == nil || namespace == "" || secretName == "" {
		return
	}

	backoff := 1 * time.Second
	const maxBackoff = 15 * time.Second

	for {
		select {
		case <-ctx.Done():
			log.Println("[TLS] Secret watcher stopped")
			return
		default:
		}

		watcher, err := kubeClient.CoreV1().Secrets(namespace).Watch(ctx, metav1.ListOptions{
			FieldSelector: "metadata.name=" + secretName,
		})
		if err != nil {
			log.Printf("[TLS] Error starting watch on Secret %s/%s: %v; retrying in %v", namespace, secretName, err, backoff)
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				continue
			}
		}

		backoff = 1 * time.Second // reset backoff on successful watch establishment
		r.processWatchEvents(ctx, watcher, namespace, secretName)
	}
}

// processWatchEvents processes incoming Secret watch events.
func (r *CertReloader) processWatchEvents(ctx context.Context, watcher watch.Interface, namespace, secretName string) {
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.ResultChan():
			if !ok {
				log.Printf("[TLS] Secret watch channel closed for %s/%s; will reconnect", namespace, secretName)
				return
			}

			switch event.Type {
			case watch.Added, watch.Modified:
				secret, ok := event.Object.(*corev1.Secret)
				if !ok {
					continue
				}

				certPEM := secret.Data["tls.crt"]
				keyPEM := secret.Data["tls.key"]
				if len(certPEM) == 0 || len(keyPEM) == 0 {
					log.Printf("[TLS] Warning: Secret %s/%s update missing tls.crt or tls.key; retaining last known-good certificate", namespace, secretName)
					continue
				}

				if err := r.ReloadFromBytes(certPEM, keyPEM); err != nil {
					log.Printf("[TLS] Warning: Failed to reload certificate from Secret %s/%s: %v; retaining last known-good certificate", namespace, secretName, err)
				}

			case watch.Deleted:
				log.Printf("[TLS] Warning: Configured TLS Secret %s/%s was deleted; retaining last known-good certificate", namespace, secretName)

			case watch.Error:
				log.Printf("[TLS] Secret watch received error for %s/%s; will reconnect", namespace, secretName)
				return
			}
		}
	}
}
