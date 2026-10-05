package reliability

import (
	"testing"

	"github.com/Burhan-21/kubeguard/internal/normalizer"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }

func makeRelResource(kind string, containers []corev1.Container, podSpecMod func(*corev1.PodSpec)) *normalizer.NormalizedResource {
	spec := corev1.PodSpec{
		Containers: containers,
	}
	if podSpecMod != nil {
		podSpecMod(&spec)
	}
	return &normalizer.NormalizedResource{
		Kind:    kind,
		Name:    "reliability-workload",
		PodSpec: &spec,
	}
}

func TestKG_REL_001_ReadinessProbe(t *testing.T) {
	rule := &ReadinessProbeRule{}

	// Positive test: missing readiness probe
	posRes := makeRelResource("Deployment", []corev1.Container{{Name: "app"}}, nil)
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-REL-001" {
		t.Errorf("KG-REL-001 positive test failed, got %+v", fPos)
	}

	// Negative test: readiness probe defined
	negRes := makeRelResource("Deployment", []corev1.Container{
		{
			Name:           "app",
			ReadinessProbe: &corev1.Probe{},
		},
	}, nil)
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-REL-001 negative test failed, got %+v", fNeg)
	}
}

func TestKG_REL_002_LivenessProbe(t *testing.T) {
	rule := &LivenessProbeRule{}

	// Positive test: missing liveness probe
	posRes := makeRelResource("Deployment", []corev1.Container{{Name: "app"}}, nil)
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-REL-002" {
		t.Errorf("KG-REL-002 positive test failed, got %+v", fPos)
	}

	// Negative test: liveness probe defined
	negRes := makeRelResource("Deployment", []corev1.Container{
		{
			Name:          "app",
			LivenessProbe: &corev1.Probe{},
		},
	}, nil)
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-REL-002 negative test failed, got %+v", fNeg)
	}
}

func TestKG_REL_003_StartupProbe(t *testing.T) {
	rule := &StartupProbeRule{}

	// Positive test: has liveness probe but no startup probe
	posRes := makeRelResource("Deployment", []corev1.Container{
		{
			Name:          "app",
			LivenessProbe: &corev1.Probe{},
		},
	}, nil)
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-REL-003" {
		t.Errorf("KG-REL-003 positive test failed, got %+v", fPos)
	}

	// Negative test: startup probe defined
	negRes := makeRelResource("Deployment", []corev1.Container{
		{
			Name:          "app",
			LivenessProbe: &corev1.Probe{},
			StartupProbe:  &corev1.Probe{},
		},
	}, nil)
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-REL-003 negative test failed, got %+v", fNeg)
	}
}

func TestKG_REL_004_ResourceRequests(t *testing.T) {
	rule := &ResourceRequestsRule{}

	// Positive test: missing resource requests
	posRes := makeRelResource("Deployment", []corev1.Container{{Name: "app"}}, nil)
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-REL-004" {
		t.Errorf("KG-REL-004 positive test failed, got %+v", fPos)
	}

	// Negative test: CPU and memory requests defined
	negRes := makeRelResource("Deployment", []corev1.Container{
		{
			Name: "app",
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
		},
	}, nil)
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-REL-004 negative test failed, got %+v", fNeg)
	}
}

func TestKG_REL_005_ResourceLimits(t *testing.T) {
	rule := &ResourceLimitsRule{}

	// Positive test: missing limits
	posRes := makeRelResource("Deployment", []corev1.Container{{Name: "app"}}, nil)
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-REL-005" {
		t.Errorf("KG-REL-005 positive test failed, got %+v", fPos)
	}

	// Negative test: CPU and memory limits defined
	negRes := makeRelResource("Deployment", []corev1.Container{
		{
			Name: "app",
			Resources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("512Mi"),
				},
			},
		},
	}, nil)
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-REL-005 negative test failed, got %+v", fNeg)
	}
}

func TestKG_REL_006_SingleReplica(t *testing.T) {
	rule := &SingleReplicaRule{}

	// Positive test: 1 replica
	posRes := makeRelResource("Deployment", []corev1.Container{{Name: "app"}}, nil)
	posRes.Replicas = int32Ptr(1)
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-REL-006" {
		t.Errorf("KG-REL-006 positive test failed, got %+v", fPos)
	}

	// Negative test: 3 replicas
	negRes := makeRelResource("Deployment", []corev1.Container{{Name: "app"}}, nil)
	negRes.Replicas = int32Ptr(3)
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-REL-006 negative test failed, got %+v", fNeg)
	}
}

func TestKG_REL_007_DeploymentStrategy(t *testing.T) {
	rule := &DeploymentStrategyRule{}

	// Positive test: Recreate strategy
	posRes := makeRelResource("Deployment", []corev1.Container{{Name: "app"}}, nil)
	posRes.Strategy = &appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType}
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-REL-007" {
		t.Errorf("KG-REL-007 positive test failed, got %+v", fPos)
	}

	// Negative test: RollingUpdate strategy
	negRes := makeRelResource("Deployment", []corev1.Container{{Name: "app"}}, nil)
	negRes.Strategy = &appsv1.DeploymentStrategy{Type: appsv1.RollingUpdateDeploymentStrategyType}
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-REL-007 negative test failed, got %+v", fNeg)
	}
}

func TestKG_REL_008_PDB(t *testing.T) {
	rule := &PDBRule{}

	// Positive test: Deployment with replicas > 1
	posRes := makeRelResource("Deployment", []corev1.Container{{Name: "app"}}, nil)
	posRes.Replicas = int32Ptr(2)
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-REL-008" {
		t.Errorf("KG-REL-008 positive test failed, got %+v", fPos)
	}

	// Negative test: Deployment with 1 replica (PDB rule only flags replicas > 1)
	negRes := makeRelResource("Deployment", []corev1.Container{{Name: "app"}}, nil)
	negRes.Replicas = int32Ptr(1)
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-REL-008 negative test failed, got %+v", fNeg)
	}
}

func TestKG_REL_009_Topology(t *testing.T) {
	rule := &TopologyRule{}

	// Positive test: multiple replicas and no affinity
	posRes := makeRelResource("Deployment", []corev1.Container{{Name: "app"}}, nil)
	posRes.Replicas = int32Ptr(3)
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-REL-009" {
		t.Errorf("KG-REL-009 positive test failed, got %+v", fPos)
	}

	// Negative test: affinity defined
	negRes := makeRelResource("Deployment", []corev1.Container{{Name: "app"}}, func(spec *corev1.PodSpec) {
		spec.Affinity = &corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{},
		}
	})
	negRes.Replicas = int32Ptr(3)
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-REL-009 negative test failed, got %+v", fNeg)
	}
}

func TestKG_REL_010_Termination(t *testing.T) {
	rule := &TerminationRule{}

	// Positive test: grace period is 0
	posRes := makeRelResource("Deployment", []corev1.Container{{Name: "app"}}, func(spec *corev1.PodSpec) {
		spec.TerminationGracePeriodSeconds = int64Ptr(0)
	})
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-REL-010" {
		t.Errorf("KG-REL-010 positive test failed, got %+v", fPos)
	}

	// Negative test: grace period is 30
	negRes := makeRelResource("Deployment", []corev1.Container{{Name: "app"}}, func(spec *corev1.PodSpec) {
		spec.TerminationGracePeriodSeconds = int64Ptr(30)
	})
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-REL-010 negative test failed, got %+v", fNeg)
	}
}

func TestKG_REL_011_NetworkPolicy(t *testing.T) {
	rule := &NetworkPolicyRule{}

	// Positive test: Deployment triggers advisory
	posRes := makeRelResource("Deployment", []corev1.Container{{Name: "app"}}, nil)
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-REL-011" {
		t.Errorf("KG-REL-011 positive test failed, got %+v", fPos)
	}

	// Negative test: ConfigMap does not trigger NetworkPolicy rule
	negRes := &normalizer.NormalizedResource{
		Kind: "ConfigMap",
		Name: "my-config",
	}
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-REL-011 negative test failed, got %+v", fNeg)
	}
}

func TestKG_REL_012_ServiceExposure(t *testing.T) {
	rule := &ServiceExposureRule{}

	// Positive test: Service type LoadBalancer
	svcObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name": "public-svc",
			},
			"spec": map[string]interface{}{
				"type": "LoadBalancer",
			},
		},
	}
	posRes := &normalizer.NormalizedResource{
		Kind: "Service",
		Name: "public-svc",
		Raw:  svcObj,
	}
	fPos := rule.Evaluate(posRes)
	if len(fPos) == 0 || fPos[0].RuleID != "KG-REL-012" {
		t.Errorf("KG-REL-012 positive test failed, got %+v", fPos)
	}

	// Negative test: Service type ClusterIP
	cleanSvcObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name": "internal-svc",
			},
			"spec": map[string]interface{}{
				"type": "ClusterIP",
			},
		},
	}
	negRes := &normalizer.NormalizedResource{
		Kind: "Service",
		Name: "internal-svc",
		Raw:  cleanSvcObj,
	}
	fNeg := rule.Evaluate(negRes)
	if len(fNeg) != 0 {
		t.Errorf("KG-REL-012 negative test failed, got %+v", fNeg)
	}
}

func TestAllReliabilityRulesCount(t *testing.T) {
	rules := AllReliabilityRules()
	if len(rules) != 12 {
		t.Fatalf("expected 12 registered reliability rules, got %d", len(rules))
	}
}
