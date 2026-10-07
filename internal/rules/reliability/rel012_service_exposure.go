package reliability

import (
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

type ServiceExposureRule struct{}

func (r *ServiceExposureRule) ID() string    { return "KG-REL-012" }
func (r *ServiceExposureRule) Title() string { return "External Service Exposure" }
func (r *ServiceExposureRule) Description() string {
	return "Services exposed externally should be intentional"
}
func (r *ServiceExposureRule) Category() string                   { return "reliability" }
func (r *ServiceExposureRule) DefaultSeverity() findings.Severity { return findings.SeverityWarn }

func (r *ServiceExposureRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.Raw == nil || res.Raw.GetKind() != "Service" {
		return nil
	}

	spec, ok, _ := k8sUnstructuredNestedMap(res.Raw.Object, "spec")
	if !ok {
		return nil
	}

	typeVal, _ := spec["type"].(string)
	if typeVal == "LoadBalancer" || typeVal == "NodePort" {
		return []findings.Finding{
			newFinding(r, res, "", "spec.type", "Service is externally exposed.", "Exposing services might unintentionally grant access to sensitive data.", "Verify that external exposure is intended."),
		}
	}
	return nil
}

func k8sUnstructuredNestedMap(obj map[string]interface{}, fields ...string) (map[string]interface{}, bool, error) {
	val, ok := obj[fields[0]]
	if !ok {
		return nil, false, nil
	}
	m, ok := val.(map[string]interface{})
	return m, ok, nil
}
