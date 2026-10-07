package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestRootCommandsRegistration(t *testing.T) {
	commands := rootCmd.Commands()
	expected := map[string]bool{
		"version":   false,
		"scan":      false,
		"policy":    false,
		"admission": false,
	}

	for _, c := range commands {
		if _, ok := expected[c.Name()]; ok {
			expected[c.Name()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("expected subcommand %q to be registered on rootCmd", name)
		}
	}
}

func TestVersionCommandOutput(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	versionCmd.Run(versionCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "KubeGuard Version:") {
		t.Errorf("expected version output to contain 'KubeGuard Version:', got: %s", out)
	}
}

func TestPolicyListCommandOutput(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	listCmd.Run(listCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "KG-SEC-001") {
		t.Errorf("expected policy list to contain KG-SEC-001, got: %s", out)
	}
	if !strings.Contains(out, "KG-REL-001") {
		t.Errorf("expected policy list to contain KG-REL-001, got: %s", out)
	}
}
