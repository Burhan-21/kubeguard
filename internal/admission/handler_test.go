package admission

import (
	"encoding/json"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/Burhan-21/kubeguard/internal/policy"
	"github.com/Burhan-21/kubeguard/internal/rules/security"
)

func TestAdmissionHandlerEnforceBlock(t *testing.T) {
	// Privileged deployment JSON
	rawDeploy := []byte(`{
		"apiVersion": "apps/v1",
		"kind": "Deployment",
		"metadata": {"name": "insecure-deploy", "namespace": "default"},
		"spec": {
			"replicas": 1,
			"template": {
				"spec": {
					"containers": [{
						"name": "root-container",
						"image": "nginx:1.25.0",
						"securityContext": {"privileged": true}
					}]
				}
			}
		}
	}`)

	engine := policy.NewEngine(security.AllSecurityRules())
	handler := &Handler{
		Engine: engine,
		Mode:   "enforce",
	}

	review := &admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID: "req-12345",
			Object: runtime.RawExtension{
				Raw: rawDeploy,
			},
		},
	}

	resp := handler.Handle(review)
	if resp.Allowed {
		t.Errorf("expected privileged deployment to be blocked in enforce mode")
	}
	if resp.UID != "req-12345" {
		t.Errorf("expected UID req-12345, got %s", resp.UID)
	}
	if resp.Result == nil || resp.Result.Message == "" {
		t.Errorf("expected denial message in status result")
	}
}

func TestAdmissionHandlerAuditMode(t *testing.T) {
	rawDeploy := []byte(`{
		"apiVersion": "apps/v1",
		"kind": "Deployment",
		"metadata": {"name": "insecure-deploy", "namespace": "default"},
		"spec": {
			"replicas": 1,
			"template": {
				"spec": {
					"containers": [{
						"name": "root-container",
						"image": "nginx:1.25.0",
						"securityContext": {"privileged": true}
					}]
				}
			}
		}
	}`)

	engine := policy.NewEngine(security.AllSecurityRules())
	handler := &Handler{
		Engine: engine,
		Mode:   "audit",
	}

	review := &admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID: "req-audit-1",
			Object: runtime.RawExtension{
				Raw: rawDeploy,
			},
		},
	}

	resp := handler.Handle(review)
	if !resp.Allowed {
		t.Errorf("expected audit mode to allow deployment even with violations")
	}
}

func TestAdmissionHandlerWarnMode(t *testing.T) {
	rawDeploy := []byte(`{
		"apiVersion": "apps/v1",
		"kind": "Deployment",
		"metadata": {"name": "insecure-deploy", "namespace": "default"},
		"spec": {
			"replicas": 1,
			"template": {
				"spec": {
					"containers": [{
						"name": "root-container",
						"image": "nginx:1.25.0",
						"securityContext": {"privileged": true}
					}]
				}
			}
		}
	}`)

	engine := policy.NewEngine(security.AllSecurityRules())
	handler := &Handler{
		Engine: engine,
		Mode:   "warn",
	}

	review := &admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID: "req-warn-1",
			Object: runtime.RawExtension{
				Raw: rawDeploy,
			},
		},
	}

	resp := handler.Handle(review)
	if !resp.Allowed {
		t.Errorf("expected warn mode to allow deployment")
	}
	if len(resp.Warnings) == 0 {
		t.Errorf("expected warnings in response")
	}
}

func TestAdmissionHandlerCleanWorkload(t *testing.T) {
	cleanDeploy := []byte(`{
		"apiVersion": "apps/v1",
		"kind": "Deployment",
		"metadata": {"name": "clean-deploy", "namespace": "prod"},
		"spec": {
			"replicas": 3,
			"template": {
				"spec": {
					"automountServiceAccountToken": false,
					"containers": [{
						"name": "web",
						"image": "nginx:1.25.0@sha256:104c7c5c54f2685f0f4f33d607e60b12",
						"securityContext": {
							"privileged": false,
							"allowPrivilegeEscalation": false,
							"runAsNonRoot": true,
							"runAsUser": 10001
						}
					}]
				}
			}
		}
	}`)

	engine := policy.NewEngine(security.AllSecurityRules())
	handler := &Handler{
		Engine: engine,
		Mode:   "enforce",
	}

	review := &admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID: "req-clean-1",
			Object: runtime.RawExtension{
				Raw: cleanDeploy,
			},
		},
	}

	resp := handler.Handle(review)
	if !resp.Allowed {
		t.Errorf("expected clean workload to be allowed")
	}
}

func TestAdmissionHandlerNilObject(t *testing.T) {
	engine := policy.NewEngine(security.AllSecurityRules())
	handler := &Handler{Engine: engine, Mode: "enforce"}
	review := &admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID: "req-nil",
		},
	}
	resp := handler.Handle(review)
	if !resp.Allowed {
		t.Errorf("expected nil object to default to allowed")
	}
}

// Unused dummy import verification
var _ = json.Unmarshal
var _ = metav1.Now
