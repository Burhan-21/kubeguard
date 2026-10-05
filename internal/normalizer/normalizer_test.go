package normalizer

import (
	"strings"
	"testing"

	"github.com/Burhan-21/kubeguard/internal/parser"
)

func TestNormalizeDeployment(t *testing.T) {
	yamlDoc := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
  namespace: prod
  labels:
    app: web
spec:
  replicas: 3
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
      - name: nginx
        image: nginx:1.25.0
`
	objs, err := parser.ParseReader(strings.NewReader(yamlDoc))
	if err != nil || len(objs) == 0 {
		t.Fatalf("failed to parse test deployment: %v", err)
	}

	res, err := Normalize(objs[0])
	if err != nil {
		t.Fatalf("Normalize failed: %v", err)
	}

	if res.Kind != "Deployment" {
		t.Errorf("expected Kind Deployment, got %s", res.Kind)
	}
	if res.Name != "web-app" {
		t.Errorf("expected Name web-app, got %s", res.Name)
	}
	if res.Namespace != "prod" {
		t.Errorf("expected Namespace prod, got %s", res.Namespace)
	}
	if res.Replicas == nil || *res.Replicas != 3 {
		t.Errorf("expected replicas 3, got %v", res.Replicas)
	}
	if res.PodSpec == nil || len(res.PodSpec.Containers) != 1 {
		t.Fatalf("expected 1 container in normalized PodSpec")
	}
	if res.PodSpec.Containers[0].Name != "nginx" {
		t.Errorf("expected container name nginx, got %s", res.PodSpec.Containers[0].Name)
	}
}

func TestNormalizePod(t *testing.T) {
	yamlDoc := `
apiVersion: v1
kind: Pod
metadata:
  name: standalone-pod
spec:
  containers:
  - name: busybox
    image: busybox:1.36.1
`
	objs, err := parser.ParseReader(strings.NewReader(yamlDoc))
	if err != nil || len(objs) == 0 {
		t.Fatalf("failed to parse test pod: %v", err)
	}

	res, err := Normalize(objs[0])
	if err != nil {
		t.Fatalf("Normalize failed: %v", err)
	}

	if res.Kind != "Pod" {
		t.Errorf("expected Kind Pod, got %s", res.Kind)
	}
	if res.PodSpec == nil || len(res.PodSpec.Containers) != 1 {
		t.Fatalf("expected 1 container in normalized PodSpec")
	}
}

func TestNormalizeStatefulSetAndDaemonSet(t *testing.T) {
	yamlDoc := `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: stateful-db
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: db
        image: postgres:16
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluentd-agent
spec:
  template:
    spec:
      containers:
      - name: fluentd
        image: fluentd:v1.16
`
	objs, err := parser.ParseReader(strings.NewReader(yamlDoc))
	if err != nil || len(objs) != 2 {
		t.Fatalf("failed to parse multi-doc: %v", err)
	}

	resSts, err := Normalize(objs[0])
	if err != nil || resSts.Kind != "StatefulSet" || resSts.PodSpec == nil {
		t.Errorf("StatefulSet normalization failed: %v", err)
	}

	resDs, err := Normalize(objs[1])
	if err != nil || resDs.Kind != "DaemonSet" || resDs.PodSpec == nil {
		t.Errorf("DaemonSet normalization failed: %v", err)
	}
}

func TestNormalizeNonWorkload(t *testing.T) {
	yamlDoc := `
apiVersion: v1
kind: Service
metadata:
  name: web-service
spec:
  type: ClusterIP
`
	objs, err := parser.ParseReader(strings.NewReader(yamlDoc))
	if err != nil || len(objs) == 0 {
		t.Fatalf("failed to parse: %v", err)
	}

	res, err := Normalize(objs[0])
	if err != nil {
		t.Fatalf("Normalize should succeed for non-workload: %v", err)
	}
	if res.Kind != "Service" {
		t.Errorf("expected Kind Service, got %s", res.Kind)
	}
	if res.PodSpec != nil {
		t.Errorf("expected nil PodSpec for Service, got %v", res.PodSpec)
	}
}
