package security

import (
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

type ExcessiveRBACRule struct{}

func (r *ExcessiveRBACRule) ID() string { return "KG-SEC-014" }
func (r *ExcessiveRBACRule) Title() string { return "Excessive RBAC Permissions" }
func (r *ExcessiveRBACRule) Description() string { return "Roles should not grant excessive permissions" }
func (r *ExcessiveRBACRule) Category() string { return "security" }
func (r *ExcessiveRBACRule) DefaultSeverity() findings.Severity { return findings.SeverityBlock }

func (r *ExcessiveRBACRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.Raw == nil { return nil }
	kind := res.Raw.GetKind()
	if kind != "ClusterRole" && kind != "Role" { return nil }
	
	var results []findings.Finding
	rules, ok, err := k8sUnstructuredNestedSlice(res.Raw.Object, "rules")
	if !ok || err != nil { return nil }
	
	for i, ruleRaw := range rules {
		ruleMap, ok := ruleRaw.(map[string]interface{})
		if !ok { continue }
		
		resourcesHasStar := containsStringAny(ruleMap["resources"], "*")
		verbsHasStar := containsStringAny(ruleMap["verbs"], "*")
		
		if resourcesHasStar && verbsHasStar {
			results = append(results, newFinding(r, res, "", "", "Role grants excessive permissions.", "Wildcard permissions allow unrestricted access to cluster resources.", "Use least privilege principle for RBAC."))
			_ = i // to avoid unused error, we just report per role, though we could specify index
		}
	}
	
	return results
}

func k8sUnstructuredNestedSlice(obj map[string]interface{}, fields ...string) ([]interface{}, bool, error) {
	val, ok := obj[fields[0]]
	if !ok { return nil, false, nil }
	slice, ok := val.([]interface{})
	return slice, ok, nil
}

func containsStringAny(val interface{}, target string) bool {
	if slice, ok := val.([]interface{}); ok {
		for _, item := range slice {
			if str, ok := item.(string); ok && str == target {
				return true
			}
		}
	}
	return false
}
