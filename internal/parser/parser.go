package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Parse reads a single YAML file and returns unstructured objects.
// Supports multi-document YAML (--- separated).
func Parse(path string) ([]unstructured.Unstructured, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer f.Close()

	return ParseReader(f)
}

// ParseReader reads YAML from an io.Reader.
func ParseReader(r io.Reader) ([]unstructured.Unstructured, error) {
	var resources []unstructured.Unstructured

	decoder := yaml.NewYAMLReader(bufio.NewReader(r))
	for {
		doc, err := decoder.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading yaml document: %w", err)
		}

		doc = bytes.TrimSpace(doc)
		if len(doc) == 0 {
			continue
		}

		obj := unstructured.Unstructured{}
		if err := yaml.Unmarshal(doc, &obj); err != nil {
			return nil, fmt.Errorf("error unmarshaling yaml document: %w", err)
		}

		if len(obj.Object) == 0 {
			continue
		}

		resources = append(resources, obj)
	}

	return resources, nil
}

// ParseDirectory reads all .yaml/.yml files from a directory.
func ParseDirectory(dir string) ([]unstructured.Unstructured, error) {
	var allResources []unstructured.Unstructured

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if ext == ".yaml" || ext == ".yml" {
			path := filepath.Join(dir, entry.Name())
			resources, err := Parse(path)
			if err != nil {
				return nil, fmt.Errorf("failed to parse file %s: %w", path, err)
			}
			allResources = append(allResources, resources...)
		}
	}

	return allResources, nil
}
