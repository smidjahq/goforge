package config

// Config holds the fully-resolved choices made by the user, either through
// the interactive TUI prompt or via CLI flags. All downstream packages
// (validator, generator, postgen) consume this struct.
type Config struct {
	Name      string   // project directory name, e.g. "myapp"
	Module    string   // Go module path, e.g. "github.com/you/myapp"
	Framework string   // gin | chi | echo | fiber
	DB        string   // postgres-gorm | postgres-sqlc | postgres-raw | sqlite-gorm | sqlite-raw | mysql-gorm | mysql-sqlc | mysql-raw | none
	Logger    string   // slog | zap | zerolog
	Extras    []string // any of: docker, makefile, ci, swagger, migrations, linter
	GoVersion string   // Go version for go.mod and Dockerfile, e.g. "1.26"
}
