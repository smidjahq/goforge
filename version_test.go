package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCmd_ContainsGoforge(t *testing.T) {
	var buf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "goforge") {
		t.Errorf("version output %q missing \"goforge\"", out)
	}
}

func TestVersionCmd_ContainsAllFields(t *testing.T) {
	var buf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"commit:", "built:", "go:"} {
		if !strings.Contains(out, want) {
			t.Errorf("version output %q missing field %q", out, want)
		}
	}
}

func TestVersionCmd_DefaultsWhenNoLdflags(t *testing.T) {
	// When built without -ldflags, version/commit/date retain their zero values.
	var buf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	// At least one of the default placeholders must appear in the output,
	// confirming the command runs without injected build metadata.
	hasDefault := strings.Contains(out, "dev") ||
		strings.Contains(out, "none") ||
		strings.Contains(out, "unknown")
	if !hasDefault {
		t.Errorf("expected default ldflag placeholder in output, got: %q", out)
	}
}

func TestVersionCmd_ContainsGoRuntime(t *testing.T) {
	var buf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// runtime.Version() always starts with "go"
	if !strings.Contains(buf.String(), "go") {
		t.Errorf("version output missing Go runtime version")
	}
}
