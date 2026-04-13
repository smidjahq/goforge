package main

import (
	"bytes"
	"strings"
	"testing"
)

func runList(t *testing.T, args ...string) (string, error) {
	t.Helper()
	buf := &bytes.Buffer{}
	root := newRootCmd()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestListCmd_Frameworks(t *testing.T) {
	out, err := runList(t, "list", "frameworks")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"gin", "chi"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q; got: %q", want, out)
		}
	}
}

func TestListCmd_DBs(t *testing.T) {
	out, err := runList(t, "list", "dbs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"postgres-gorm", "mysql-gorm", "sqlite-raw", "none"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q; got: %q", want, out)
		}
	}
}

func TestListCmd_Loggers(t *testing.T) {
	out, err := runList(t, "list", "loggers")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"slog", "zap", "zerolog"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q; got: %q", want, out)
		}
	}
}

func TestListCmd_Extras(t *testing.T) {
	out, err := runList(t, "list", "extras")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"docker", "migrations"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q; got: %q", want, out)
		}
	}
}

func TestListCmd_Presets(t *testing.T) {
	out, err := runList(t, "list", "presets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(strings.ToLower(out), "no presets") {
		t.Errorf("expected 'no presets' message, got: %q", out)
	}
}

func TestListCmd_UnknownCategory(t *testing.T) {
	out, err := runList(t, "list", "foobar")
	combined := out
	if err != nil {
		combined += err.Error()
	}
	if !strings.Contains(combined, "unknown category") {
		t.Errorf("expected 'unknown category' in output/error, got: %q", combined)
	}
	if !strings.Contains(combined, "foobar") {
		t.Errorf("expected 'foobar' echoed back in output/error, got: %q", combined)
	}
}

func TestListCmd_CaseInsensitive(t *testing.T) {
	out, err := runList(t, "list", "FRAMEWORKS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "gin") {
		t.Errorf("case-insensitive lookup for FRAMEWORKS missing 'gin'; got: %q", out)
	}
}

func TestListCmd_EachItemOnOwnLine(t *testing.T) {
	out, err := runList(t, "list", "loggers")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.ContainsAny(line, ", ") {
			t.Errorf("line %q appears to contain multiple items (expected one per line)", line)
		}
	}
}