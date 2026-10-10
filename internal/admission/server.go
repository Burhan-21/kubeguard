package admission

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/client-go/kubernetes"
)

type Server struct {
	Handler         *Handler
	CertFile        string
	KeyFile         string
	Port            int
	SecretName      string
	SecretNamespace string
	KubeClient      kubernetes.Interface
	ReloadInterval  time.Duration
	Reloader        *CertReloader
	httpServer      *http.Server
	cancelWatch     context.CancelFunc
}

func (s *Server) Start() error {
	return s.StartWithContext(context.Background())
}

func (s *Server) StartWithContext(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", s.serveValidate)
	mux.HandleFunc("/healthz", s.serveHealthz)
	mux.HandleFunc("/readyz", s.serveReadyz)
	mux.HandleFunc("/metrics", s.serveMetrics)

	addr := fmt.Sprintf(":%d", s.Port)

	if s.CertFile != "" && s.KeyFile != "" {
		if s.Reloader == nil {
			reloader, err := NewCertReloader(s.CertFile, s.KeyFile)
			if err != nil {
				return fmt.Errorf("failed to load initial TLS certificate: %w", err)
			}
			s.Reloader = reloader
		}

		watchCtx, cancel := context.WithCancel(ctx)
		s.cancelWatch = cancel

		// Start background file watcher
		reloadInterval := s.ReloadInterval
		if reloadInterval <= 0 {
			reloadInterval = 1 * time.Second
		}
		go s.Reloader.StartFileWatcher(watchCtx, reloadInterval)

		// Start background Kubernetes Secret watcher if client and secret are configured
		if s.KubeClient != nil && s.SecretName != "" && s.SecretNamespace != "" {
			go s.Reloader.StartSecretWatcher(watchCtx, s.KubeClient, s.SecretNamespace, s.SecretName)
		}

		s.httpServer = &http.Server{
			Addr:    addr,
			Handler: mux,
			TLSConfig: &tls.Config{
				GetCertificate: s.Reloader.GetCertificate,
				MinVersion:     tls.VersionTLS12,
			},
		}

		return s.httpServer.ListenAndServeTLS("", "")
	}

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.cancelWatch != nil {
		s.cancelWatch()
	}
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

func (s *Server) serveHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) serveReadyz(w http.ResponseWriter, _ *http.Request) {
	// If TLS is configured, ensure a certificate is actively loaded
	if s.Reloader != nil {
		if _, _, _, err := s.Reloader.CurrentCertInfo(); err != nil {
			http.Error(w, "TLS certificate not ready", http.StatusServiceUnavailable)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) serveMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	DefaultMetrics().WritePrometheus(w)
}

func (s *Server) serveValidate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	DefaultMetrics().IncRequestsTotal()

	var body []byte
	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		DefaultMetrics().IncRequestErrorsTotal()
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	var review admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &review); err != nil {
		DefaultMetrics().IncRequestErrorsTotal()
		http.Error(w, fmt.Sprintf("could not parse request: %v", err), http.StatusBadRequest)
		return
	}

	if review.Request == nil {
		DefaultMetrics().IncRequestErrorsTotal()
		http.Error(w, "admission review request is nil", http.StatusBadRequest)
		return
	}

	response := s.Handler.Handle(&review)
	if response != nil {
		if response.Allowed {
			DefaultMetrics().IncRequestsAllowed()
			if len(response.Warnings) > 0 {
				DefaultMetrics().IncRequestsWarned()
			}
		} else {
			DefaultMetrics().IncRequestsDenied()
		}
	}

	review.Response = response
	if review.APIVersion == "" {
		review.APIVersion = "admission.k8s.io/v1"
	}
	if review.Kind == "" {
		review.Kind = "AdmissionReview"
	}

	respBytes, err := json.Marshal(review)
	if err != nil {
		DefaultMetrics().IncPolicyErrorsTotal()
		http.Error(w, fmt.Sprintf("could not marshal response: %v", err), http.StatusInternalServerError)
		return
	}

	DefaultMetrics().ObserveRequestDuration(time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(respBytes)
}
