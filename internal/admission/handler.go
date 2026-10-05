package admission

import (
	"encoding/json"
	"fmt"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
	"github.com/Burhan-21/kubeguard/internal/policy"
)

type Handler struct {
	Engine *policy.Engine
	Mode   string // audit, warn, enforce
}

func (h *Handler) Handle(review *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	req := review.Request
	resp := &admissionv1.AdmissionResponse{
		UID:     req.UID,
		Allowed: true,
	}

	if req.Object.Raw == nil {
		return resp
	}

	var obj unstructured.Unstructured
	if err := json.Unmarshal(req.Object.Raw, &obj); err != nil {
		resp.Result = &metav1.Status{
			Message: fmt.Sprintf("failed to parse object: %v", err),
		}
		return resp
	}

	norm, err := normalizer.Normalize(obj)
	if err != nil {
		// Ignore resources we don't support normalizing
		return resp
	}

	scanResult := h.Engine.EvaluateAll([]*normalizer.NormalizedResource{norm})
	
	var blockMsgs []string
	var warnMsgs []string

	for _, f := range scanResult.Findings {
		if f.Severity == findings.SeverityBlock {
			blockMsgs = append(blockMsgs, fmt.Sprintf("[%s] %s", f.RuleID, f.Message))
		} else if f.Severity == findings.SeverityWarn {
			warnMsgs = append(warnMsgs, fmt.Sprintf("[%s] %s", f.RuleID, f.Message))
		}
	}

	switch h.Mode {
	case "enforce":
		if len(blockMsgs) > 0 {
			resp.Allowed = false
			resp.Result = &metav1.Status{
				Message: fmt.Sprintf("KubeGuard block: %s", strings.Join(blockMsgs, "; ")),
			}
		} else if len(warnMsgs) > 0 {
			resp.Warnings = warnMsgs
		}
	case "warn":
		allWarnings := append(blockMsgs, warnMsgs...)
		if len(allWarnings) > 0 {
			resp.Warnings = allWarnings
		}
	case "audit":
		// Do nothing to the response, log findings internally
	}

	return resp
}
