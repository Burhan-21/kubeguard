package admission

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
)

type Server struct {
	Handler  *Handler
	CertFile string
	KeyFile  string
	Port     int
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", s.serveValidate)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }) // Stub for metrics

	addr := fmt.Sprintf(":%d", s.Port)
	if s.CertFile != "" && s.KeyFile != "" {
		return http.ListenAndServeTLS(addr, s.CertFile, s.KeyFile, mux)
	}
	return http.ListenAndServe(addr, mux)
}

func (s *Server) serveValidate(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	var review admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &review); err != nil {
		http.Error(w, fmt.Sprintf("could not parse request: %v", err), http.StatusBadRequest)
		return
	}

	if review.Request == nil {
		http.Error(w, "admission review request is nil", http.StatusBadRequest)
		return
	}

	response := s.Handler.Handle(&review)

	review.Response = response
	// The response needs to match the GVK of the request
	review.SetGroupVersionKind(review.GroupVersionKind())

	respBytes, err := json.Marshal(review)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not marshal response: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
}
