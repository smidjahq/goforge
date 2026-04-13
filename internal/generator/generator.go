// Package generator walks the embedded template tree, renders each .tmpl file
// via text/template, and writes the output to disk. Non-.tmpl files are copied
// as-is. All template data is sourced from a config.Config value.
package generator

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"text/template"

	"github.com/smidjahq/goforge/internal/config"
)

//go:embed all:templates
var templateFS embed.FS

// Generate produces the project file tree under outputPath for the given cfg.
// The output directory is created if it does not exist.
// Layers are merged in order: base → framework → db → logger → extras.
// Later layers overwrite files from earlier layers, enabling framework/db-specific overrides.
// progress is called with a human-readable status message before each layer is rendered;
// pass nil to disable progress reporting.
func Generate(cfg config.Config, outputPath string, progress func(string)) error {
	if err := os.MkdirAll(outputPath, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	layers := activeLayers(cfg)
	for _, layer := range layers {
		if progress != nil {
			progress(layerMessage(layer))
		}
		if err := renderLayer(cfg, layer, outputPath); err != nil {
			return fmt.Errorf("render layer %q: %w", layer, err)
		}
	}
	return nil
}

// layerMessage returns a human-readable progress message for a template layer path.
func layerMessage(layer string) string {
	switch {
	case layer == "templates/base":
		return "Applying base structure"
	case strings.HasPrefix(layer, "templates/frameworks/"):
		fw := strings.TrimPrefix(layer, "templates/frameworks/")
		return "Adding " + fw + " framework"
	case strings.HasPrefix(layer, "templates/db/"):
		db := strings.TrimPrefix(layer, "templates/db/")
		return "Configuring " + db + " database"
	case strings.HasPrefix(layer, "templates/loggers/"):
		lg := strings.TrimPrefix(layer, "templates/loggers/")
		return "Setting up " + lg + " logger"
	case strings.HasPrefix(layer, "templates/extras/"):
		ex := strings.TrimPrefix(layer, "templates/extras/")
		return "Adding " + ex
	default:
		return "Processing " + layer
	}
}

// ListFiles returns the relative paths of all files that would be generated for
// cfg, in the order they would be written. Later layers may overwrite paths
// from earlier layers; the returned slice reflects the final de-duplicated set,
// sorted alphabetically.
func ListFiles(cfg config.Config) []string {
	seen := make(map[string]struct{})
	var ordered []string

	for _, layer := range activeLayers(cfg) {
		_ = fs.WalkDir(templateFS, layer, func(fsPath string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(filepath.FromSlash(layer), filepath.FromSlash(fsPath))
			if err != nil {
				return nil
			}
			if strings.HasSuffix(rel, ".tmpl") {
				rel = strings.TrimSuffix(rel, ".tmpl")
			}
			rel = filepath.ToSlash(rel)
			if _, exists := seen[rel]; !exists {
				seen[rel] = struct{}{}
				ordered = append(ordered, rel)
			}
			return nil
		})
	}

	sort.Strings(ordered)
	return ordered
}

// activeLayers returns the ordered list of template layer paths to apply for cfg.
// Order: base → framework → db → logger → extras.
// Later layers overwrite files from earlier layers, enabling progressive overrides.
func activeLayers(cfg config.Config) []string {
	layers := []string{"templates/base"}

	if cfg.Framework != "" {
		layers = append(layers, "templates/frameworks/"+cfg.Framework)
	}

	if cfg.DB != "" && cfg.DB != "none" {
		layers = append(layers, "templates/db/"+cfg.DB)
	}

	if cfg.Logger != "" {
		layers = append(layers, "templates/loggers/"+cfg.Logger)
	}

	for _, extra := range cfg.Extras {
		// migrations is only valid with a non-none DB (already validated upstream)
		if extra == "migrations" && (cfg.DB == "" || cfg.DB == "none") {
			continue
		}
		layers = append(layers, "templates/extras/"+extra)
	}

	return layers
}

// renderLayer walks one template layer and writes rendered files to outputPath.
func renderLayer(cfg config.Config, layerRoot, outputPath string) error {
	return fs.WalkDir(templateFS, layerRoot, func(fsPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// path relative to the layer root (e.g. "cmd/app/main.go.tmpl")
		rel, err := filepath.Rel(filepath.FromSlash(layerRoot), filepath.FromSlash(fsPath))
		if err != nil {
			return err
		}

		dst := filepath.Join(outputPath, rel)

		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}

		if strings.HasSuffix(fsPath, ".tmpl") {
			dst = strings.TrimSuffix(dst, ".tmpl")
			return renderTemplate(cfg, fsPath, dst)
		}

		return copyFile(fsPath, dst)
	})
}

// renderTemplate executes a single .tmpl file and writes the result to dst.
func renderTemplate(cfg config.Config, fsPath, dst string) error {
	raw, err := templateFS.ReadFile(fsPath)
	if err != nil {
		return fmt.Errorf("read template %q: %w", fsPath, err)
	}

	funcMap := template.FuncMap{
		// hasExtra reports whether name is present in the extras slice.
		"hasExtra": func(extras []string, name string) bool {
			return slices.Contains(extras, name)
		},
	}

	// Use the filename (without directory) as the template name.
	name := path.Base(fsPath)
	tmpl, err := template.New(name).Funcs(funcMap).Parse(string(raw))
	if err != nil {
		return fmt.Errorf("parse template %q: %w", fsPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create directory for %q: %w", dst, err)
	}

	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create file %q: %w", dst, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, cfg); err != nil {
		return fmt.Errorf("execute template %q: %w", fsPath, err)
	}
	return nil
}

// copyFile copies a non-template file from the embedded FS to dst.
func copyFile(fsPath, dst string) error {
	data, err := templateFS.ReadFile(fsPath)
	if err != nil {
		return fmt.Errorf("read file %q: %w", fsPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create directory for %q: %w", dst, err)
	}

	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return fmt.Errorf("write file %q: %w", dst, err)
	}
	return nil
}
