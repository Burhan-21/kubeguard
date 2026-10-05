package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseReaderSingle(t *testing.T) {
	yamlContent := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deploy
  namespace: default
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: test
        image: nginx:1.25.0
`
	resources, err := ParseReader(strings.NewReader(yamlContent))
	if err != nil {
		t.Fatalf("ParseReader failed: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}

	if resources[0].GetKind() != "Deployment" {
		t.Errorf("expected Kind Deployment, got %s", resources[0].GetKind())
	}
	if resources[0].GetName() != "test-deploy" {
		t.Errorf("expected Name test-deploy, got %s", resources[0].GetName())
	}
}

func TestParseReaderMultiDocument(t *testing.T) {
	multiDoc := `
apiVersion: v1
kind: Service
metadata:
  name: test-svc
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deploy
---
`
	resources, err := ParseReader(strings.NewReader(multiDoc))
	if err != nil {
		t.Fatalf("ParseReader failed on multi-doc: %v", err)
	}

	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}

	if resources[0].GetKind() != "Service" {
		t.Errorf("first resource expected Service, got %s", resources[0].GetKind())
	}
	if resources[1].GetKind() != "Deployment" {
		t.Errorf("second resource expected Deployment, got %s", resources[1].GetKind())
	}
}

func TestParseReaderEmpty(t *testing.T) {
	emptyYAML := `
# Just comments
---
---
`
	resources, err := ParseReader(strings.NewReader(emptyYAML))
	if err != nil {
		t.Fatalf("ParseReader failed on empty yaml: %v", err)
	}

	if len(resources) != 0 {
		t.Errorf("expected 0 resources for empty doc, got %d", len(resources))
	}
}

func TestParseReaderInvalid(t *testing.T) {
	invalidYAML := `
key: [invalid, unclosed array
`
	_, err := ParseReader(strings.NewReader(invalidYAML))
	if err == nil {
		t.Errorf("expected error for invalid yaml, got nil")
	}
}

func TestParseDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "kubeguard-parser-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	f1 := filepath.Join(tempDir, "deploy.yaml")
	f2 := filepath.Join(tempDir, "service.yml")
	f3 := filepath.Join(tempDir, "ignored.txt")

	os.WriteFile(f1, []byte("apiVersion: v1\nkind: Pod\nmetadata:\n  name: p1\n"), 0644)
	os.WriteFile(f2, []byte("apiVersion: v1\nkind: Service\nmetadata:\n  name: s1\n"), 0644)
	os.WriteFile(f3, []byte("plain text not yaml"), 0644)

	resources, err := ParseDirectory(tempDir)
	if err != nil {
		t.Fatalf("ParseDirectory failed: %v", err)
	}

	if len(resources) != 2 {
		t.Fatalf("expected 2 resources from directory, got %d", len(resources))
	}
}
