package normalizer

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// NormalizedResource represents a Kubernetes resource normalized for policy evaluation.
type NormalizedResource struct {
	Kind        string
	Name        string
	Namespace   string
	Replicas    *int32
	PodSpec     *corev1.PodSpec
	Strategy    *appsv1.DeploymentStrategy
	Labels      map[string]string
	Annotations map[string]string
	Raw         *unstructured.Unstructured
}

// Normalize extracts the PodSpec from supported resource types.
// Supports: Deployment, StatefulSet, DaemonSet, Pod, Job, CronJob
func Normalize(obj unstructured.Unstructured) (*NormalizedResource, error) {
	nr := &NormalizedResource{
		Kind:        obj.GetKind(),
		Name:        obj.GetName(),
		Namespace:   obj.GetNamespace(),
		Labels:      obj.GetLabels(),
		Annotations: obj.GetAnnotations(),
		Raw:         &obj,
	}

	switch nr.Kind {
	case "Pod":
		var pod corev1.Pod
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &pod); err != nil {
			return nil, fmt.Errorf("failed to convert to Pod: %w", err)
		}
		nr.PodSpec = &pod.Spec
	case "Deployment":
		var deploy appsv1.Deployment
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &deploy); err != nil {
			return nil, fmt.Errorf("failed to convert to Deployment: %w", err)
		}
		nr.PodSpec = &deploy.Spec.Template.Spec
		nr.Replicas = deploy.Spec.Replicas
		nr.Strategy = &deploy.Spec.Strategy
	case "StatefulSet":
		var sts appsv1.StatefulSet
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &sts); err != nil {
			return nil, fmt.Errorf("failed to convert to StatefulSet: %w", err)
		}
		nr.PodSpec = &sts.Spec.Template.Spec
		nr.Replicas = sts.Spec.Replicas
	case "DaemonSet":
		var ds appsv1.DaemonSet
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &ds); err != nil {
			return nil, fmt.Errorf("failed to convert to DaemonSet: %w", err)
		}
		nr.PodSpec = &ds.Spec.Template.Spec
	case "Job":
		var job batchv1.Job
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &job); err != nil {
			return nil, fmt.Errorf("failed to convert to Job: %w", err)
		}
		nr.PodSpec = &job.Spec.Template.Spec
	case "CronJob":
		var cj batchv1.CronJob
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &cj); err != nil {
			return nil, fmt.Errorf("failed to convert to CronJob: %w", err)
		}
		nr.PodSpec = &cj.Spec.JobTemplate.Spec.Template.Spec
	case "Service", "Role", "ClusterRole", "RoleBinding", "ClusterRoleBinding", "NetworkPolicy", "PodDisruptionBudget", "Ingress", "ConfigMap", "Secret", "ServiceAccount":
		// Non-workload resources normalized without PodSpec; Raw is preserved for custom rule evaluation.
		nr.PodSpec = nil
	default:
		// Generic fallback: normalize with PodSpec = nil so rules that inspect Raw can evaluate
		nr.PodSpec = nil
	}

	return nr, nil
}
