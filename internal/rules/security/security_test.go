package security

import (
	"testing"

	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func boolPtr(b bool) *bool    { return &b }
func int64Ptr(i int64) *int64 { return &i }

func makeResource(containers []corev1.Container, podSpecMod func(*corev1.PodSpec)) *normalizer.NormalizedResource {
	spec := corev1.PodSpec{
		Containers: containers,
	}
	if podSpecMod != nil {
		podSpecMod(&spec)
	}
	return &normalizer.NormalizedResource{
		Kind:    "Deployment",
		Name:    "test-workload",
		PodSpec: &spec,
	}
}

func TestKG_SEC_001_Privileged(t *testing.T) {
	rule := &PrivilegedRule{}

	// Positive test: privileged container
	posRes := makeResource([]corev1.Container{
		{
			Name: "app",
			SecurityContext: &corev1.SecurityContext{
				Privileged: boolPtr(true),
			},
		},
	}, nil)
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-SEC-001" || fPos[0].Severity != findings.SeverityBlock {
		t.Errorf("KG-SEC-001 positive test failed: expected 1 BLOCK finding, got %+v", fPos)
	}

	// Negative test: unprivileged container
	negRes := makeResource([]corev1.Container{
		{
			Name: "app",
			SecurityContext: &corev1.SecurityContext{
				Privileged: boolPtr(false),
			},
		},
	}, nil)
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-SEC-001 negative test failed: expected 0 findings, got %+v", fNeg)
	}
}

func TestKG_SEC_002_RunAsRoot(t *testing.T) {
	rule := &RunAsRootRule{}

	// Positive test: no runAsNonRoot and runAsUser is 0
	posRes := makeResource([]corev1.Container{
		{
			Name: "app",
			SecurityContext: &corev1.SecurityContext{
				RunAsUser: int64Ptr(0),
			},
		},
	}, nil)
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-SEC-002" {
		t.Errorf("KG-SEC-002 positive test failed, got %+v", fPos)
	}

	// Negative test: runAsNonRoot is true
	negRes := makeResource([]corev1.Container{
		{
			Name: "app",
			SecurityContext: &corev1.SecurityContext{
				RunAsNonRoot: boolPtr(true),
				RunAsUser:    int64Ptr(10001),
			},
		},
	}, nil)
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-SEC-002 negative test failed, got %+v", fNeg)
	}
}

func TestKG_SEC_003_PrivilegeEscalation(t *testing.T) {
	rule := &PrivilegeEscalationRule{}

	// Positive test: allowPrivilegeEscalation is true or nil
	posRes := makeResource([]corev1.Container{
		{
			Name: "app",
			SecurityContext: &corev1.SecurityContext{
				AllowPrivilegeEscalation: boolPtr(true),
			},
		},
	}, nil)
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-SEC-003" {
		t.Errorf("KG-SEC-003 positive test failed, got %+v", fPos)
	}

	// Negative test: allowPrivilegeEscalation is false
	negRes := makeResource([]corev1.Container{
		{
			Name: "app",
			SecurityContext: &corev1.SecurityContext{
				AllowPrivilegeEscalation: boolPtr(false),
			},
		},
	}, nil)
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-SEC-003 negative test failed, got %+v", fNeg)
	}
}

func TestKG_SEC_004_HostNetwork(t *testing.T) {
	rule := &HostNetworkRule{}

	// Positive test: hostNetwork = true
	posRes := makeResource([]corev1.Container{{Name: "app"}}, func(spec *corev1.PodSpec) {
		spec.HostNetwork = true
	})
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-SEC-004" {
		t.Errorf("KG-SEC-004 positive test failed, got %+v", fPos)
	}

	// Negative test: hostNetwork = false
	negRes := makeResource([]corev1.Container{{Name: "app"}}, func(spec *corev1.PodSpec) {
		spec.HostNetwork = false
	})
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-SEC-004 negative test failed, got %+v", fNeg)
	}
}

func TestKG_SEC_005_HostPIDIPC(t *testing.T) {
	rule := &HostPIDIPCRule{}

	// Positive test: hostPID = true
	posRes := makeResource([]corev1.Container{{Name: "app"}}, func(spec *corev1.PodSpec) {
		spec.HostPID = true
	})
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-SEC-005" {
		t.Errorf("KG-SEC-005 positive test failed, got %+v", fPos)
	}

	// Negative test: hostPID and hostIPC false
	negRes := makeResource([]corev1.Container{{Name: "app"}}, func(spec *corev1.PodSpec) {
		spec.HostPID = false
		spec.HostIPC = false
	})
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-SEC-005 negative test failed, got %+v", fNeg)
	}
}

func TestKG_SEC_006_HostPath(t *testing.T) {
	rule := &HostPathRule{}

	// Positive test: mounts /var/run/docker.sock
	posRes := makeResource([]corev1.Container{{Name: "app"}}, func(spec *corev1.PodSpec) {
		spec.Volumes = []corev1.Volume{
			{
				Name: "dockersock",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/var/run/docker.sock",
					},
				},
			},
		}
	})
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-SEC-006" || fPos[0].Severity != findings.SeverityBlock {
		t.Errorf("KG-SEC-006 dangerous hostPath test failed, got %+v", fPos)
	}

	// Negative test: EmptyDir volume only
	negRes := makeResource([]corev1.Container{{Name: "app"}}, func(spec *corev1.PodSpec) {
		spec.Volumes = []corev1.Volume{
			{
				Name: "temp",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		}
	})
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-SEC-006 negative test failed, got %+v", fNeg)
	}
}

func TestKG_SEC_007_Capabilities(t *testing.T) {
	rule := &CapabilitiesRule{}

	// Positive test: SYS_ADMIN capability
	posRes := makeResource([]corev1.Container{
		{
			Name: "app",
			SecurityContext: &corev1.SecurityContext{
				Capabilities: &corev1.Capabilities{
					Add: []corev1.Capability{"SYS_ADMIN"},
				},
			},
		},
	}, nil)
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-SEC-007" || fPos[0].Severity != findings.SeverityBlock {
		t.Errorf("KG-SEC-007 positive test failed, got %+v", fPos)
	}

	// Negative test: Drop ALL
	negRes := makeResource([]corev1.Container{
		{
			Name: "app",
			SecurityContext: &corev1.SecurityContext{
				Capabilities: &corev1.Capabilities{
					Drop: []corev1.Capability{"ALL"},
				},
			},
		},
	}, nil)
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-SEC-007 negative test failed, got %+v", fNeg)
	}
}

func TestKG_SEC_008_MissingSecurityContext(t *testing.T) {
	rule := &MissingSecurityContextRule{}

	// Positive test: nil security context
	posRes := makeResource([]corev1.Container{
		{
			Name:            "app",
			SecurityContext: nil,
		},
	}, nil)
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-SEC-008" {
		t.Errorf("KG-SEC-008 positive test failed, got %+v", fPos)
	}

	// Negative test: security context present
	negRes := makeResource([]corev1.Container{
		{
			Name: "app",
			SecurityContext: &corev1.SecurityContext{
				ReadOnlyRootFilesystem: boolPtr(true),
			},
		},
	}, nil)
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-SEC-008 negative test failed, got %+v", fNeg)
	}
}

func TestKG_SEC_009_LatestTag(t *testing.T) {
	rule := &LatestTagRule{}

	// Positive test: image with :latest
	posRes := makeResource([]corev1.Container{
		{Name: "app", Image: "nginx:latest"},
	}, nil)
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-SEC-009" {
		t.Errorf("KG-SEC-009 positive test failed, got %+v", fPos)
	}

	// Negative test: pinned tag
	negRes := makeResource([]corev1.Container{
		{Name: "app", Image: "nginx:1.25.3"},
	}, nil)
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-SEC-009 negative test failed, got %+v", fNeg)
	}
}

func TestKG_SEC_010_MissingTag(t *testing.T) {
	rule := &MissingTagRule{}

	// Positive test: no tag
	posRes := makeResource([]corev1.Container{
		{Name: "app", Image: "redis"},
	}, nil)
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-SEC-010" {
		t.Errorf("KG-SEC-010 positive test failed, got %+v", fPos)
	}

	// Negative test: tag present
	negRes := makeResource([]corev1.Container{
		{Name: "app", Image: "redis:7.2"},
	}, nil)
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-SEC-010 negative test failed, got %+v", fNeg)
	}
}

func TestKG_SEC_011_UnpinnedDigest(t *testing.T) {
	rule := &UnpinnedDigestRule{}

	// Positive test: image without sha256 digest
	posRes := makeResource([]corev1.Container{
		{Name: "app", Image: "nginx:1.25.0"},
	}, nil)
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-SEC-011" {
		t.Errorf("KG-SEC-011 positive test failed, got %+v", fPos)
	}

	// Negative test: image with sha256 digest
	negRes := makeResource([]corev1.Container{
		{Name: "app", Image: "nginx:1.25.0@sha256:104c7c5c54f2685f0f4f33d607e60b12"},
	}, nil)
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-SEC-011 negative test failed, got %+v", fNeg)
	}
}

func TestKG_SEC_012_EmbeddedSecret(t *testing.T) {
	rule := &EmbeddedSecretRule{}

	// Positive test: plain text password env var
	posRes := makeResource([]corev1.Container{
		{
			Name: "app",
			Env: []corev1.EnvVar{
				{Name: "DB_PASSWORD", Value: "SuperSecret123"},
			},
		},
	}, nil)
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-SEC-012" {
		t.Errorf("KG-SEC-012 positive test failed, got %+v", fPos)
	}

	// Negative test: secretKeyRef used
	negRes := makeResource([]corev1.Container{
		{
			Name: "app",
			Env: []corev1.EnvVar{
				{
					Name: "DB_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "db-secret"},
							Key:                  "password",
						},
					},
				},
			},
		},
	}, nil)
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-SEC-012 negative test failed, got %+v", fNeg)
	}
}

func TestKG_SEC_013_ServiceAccountToken(t *testing.T) {
	rule := &ServiceAccountTokenRule{}

	// Positive test: automountServiceAccountToken is true or nil
	posRes := makeResource([]corev1.Container{{Name: "app"}}, func(spec *corev1.PodSpec) {
		spec.AutomountServiceAccountToken = boolPtr(true)
	})
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-SEC-013" {
		t.Errorf("KG-SEC-013 positive test failed, got %+v", fPos)
	}

	// Negative test: automountServiceAccountToken is false
	negRes := makeResource([]corev1.Container{{Name: "app"}}, func(spec *corev1.PodSpec) {
		spec.AutomountServiceAccountToken = boolPtr(false)
	})
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-SEC-013 negative test failed, got %+v", fNeg)
	}
}

func TestKG_SEC_014_ExcessiveRBAC(t *testing.T) {
	rule := &ExcessiveRBACRule{}

	// Positive test: ClusterRole with wildcard resources and verbs
	clusterRoleObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "ClusterRole",
			"metadata": map[string]interface{}{
				"name": "cluster-admin-wildcard",
			},
			"rules": []interface{}{
				map[string]interface{}{
					"apiGroups": []interface{}{"*"},
					"resources": []interface{}{"*"},
					"verbs":     []interface{}{"*"},
				},
			},
		},
	}
	posRes := &normalizer.NormalizedResource{
		Kind: "ClusterRole",
		Name: "cluster-admin-wildcard",
		Raw:  clusterRoleObj,
	}
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-SEC-014" || fPos[0].Severity != findings.SeverityBlock {
		t.Errorf("KG-SEC-014 positive test failed, got %+v", fPos)
	}

	// Negative test: ClusterRole with restricted read permissions
	cleanRoleObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "Role",
			"metadata": map[string]interface{}{
				"name": "pod-reader",
			},
			"rules": []interface{}{
				map[string]interface{}{
					"apiGroups": []interface{}{""},
					"resources": []interface{}{"pods"},
					"verbs":     []interface{}{"get", "list"},
				},
			},
		},
	}
	negRes := &normalizer.NormalizedResource{
		Kind: "Role",
		Name: "pod-reader",
		Raw:  cleanRoleObj,
	}
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-SEC-014 negative test failed, got %+v", fNeg)
	}
}

func TestAllSecurityRulesCount(t *testing.T) {
	rules := AllSecurityRules()
	if len(rules) != 14 {
		t.Fatalf("expected 14 registered security rules, got %d", len(rules))
	}
}
