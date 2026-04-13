package prompts

import (
	"testing"
)

func TestBuildDB(t *testing.T) {
	cases := []struct {
		backend string
		layer   string
		want    string
	}{
		{"postgres", "gorm", "postgres-gorm"},
		{"postgres", "sqlc", "postgres-sqlc"},
		{"postgres", "raw", "postgres-raw"},
		{"sqlite", "gorm", "sqlite-gorm"},
		{"sqlite", "raw", "sqlite-raw"},
		{"none", "", "none"},
		// layer is ignored when backend is "none"
		{"none", "gorm", "none"},
		// empty layer falls back to backend alone
		{"postgres", "", "postgres"},
	}

	for _, tc := range cases {
		got := buildDB(tc.backend, tc.layer)
		if got != tc.want {
			t.Errorf("buildDB(%q, %q) = %q, want %q", tc.backend, tc.layer, got, tc.want)
		}
	}
}

func TestLayerOptions_Postgres(t *testing.T) {
	opts := layerOptions("postgres")
	if len(opts) == 0 {
		t.Fatal("expected options for postgres, got none")
	}
	values := make([]string, len(opts))
	for i, o := range opts {
		values[i] = o.Value
	}
	want := map[string]bool{"gorm": true, "sqlc": true, "raw": true}
	for _, v := range values {
		if !want[v] {
			t.Errorf("unexpected layer option %q for postgres", v)
		}
		delete(want, v)
	}
	for missing := range want {
		t.Errorf("missing expected layer option %q for postgres", missing)
	}
}

func TestLayerOptions_SQLite(t *testing.T) {
	opts := layerOptions("sqlite")
	if len(opts) == 0 {
		t.Fatal("expected options for sqlite, got none")
	}
	values := make(map[string]bool)
	for _, o := range opts {
		values[o.Value] = true
	}
	if !values["gorm"] {
		t.Error("sqlite should include gorm layer")
	}
	if !values["raw"] {
		t.Error("sqlite should include raw layer")
	}
	if values["sqlc"] {
		t.Error("sqlite should not include sqlc layer")
	}
}

func TestLayerOptions_None_ReturnsPlaceholder(t *testing.T) {
	opts := layerOptions("none")
	// "none" has no real layers; layerOptions returns a single placeholder option
	if len(opts) != 1 {
		t.Fatalf("expected 1 placeholder option for none, got %d", len(opts))
	}
	if opts[0].Value != "" {
		t.Errorf("placeholder option value should be empty string, got %q", opts[0].Value)
	}
}

func TestLayerOptions_UnknownBackend_ReturnsPlaceholder(t *testing.T) {
	opts := layerOptions("oracle")
	// Unknown backends behave like "none" — a placeholder is returned
	if len(opts) != 1 {
		t.Fatalf("expected 1 placeholder option for unknown backend, got %d", len(opts))
	}
}

func TestExtrasOptions_ContainsAllExtras(t *testing.T) {
	opts := extrasOptions()
	wantValues := []string{"docker", "makefile", "ci", "swagger", "migrations", "linter"}

	found := make(map[string]bool)
	for _, o := range opts {
		found[o.Value] = true
	}

	for _, want := range wantValues {
		if !found[want] {
			t.Errorf("extrasOptions missing %q", want)
		}
	}
}

func TestExtrasOptions_NoDuplicates(t *testing.T) {
	opts := extrasOptions()
	seen := make(map[string]bool)
	for _, o := range opts {
		if seen[o.Value] {
			t.Errorf("duplicate extra option value: %q", o.Value)
		}
		seen[o.Value] = true
	}
}
