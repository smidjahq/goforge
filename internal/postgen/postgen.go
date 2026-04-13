// Package postgen runs post-generation steps after the generator writes files to disk.
package postgen

import (
	"fmt"
	"os"
	"os/exec"
)

// Run executes post-generation steps in outputPath:
//   - runs `go mod tidy` (required; on failure the output directory is removed)
//   - optionally runs `git init` when initGit is true
func Run(outputPath string, initGit bool) error {
	if err := ModTidy(outputPath); err != nil {
		return err
	}
	if initGit {
		if err := GitInit(outputPath); err != nil {
			return fmt.Errorf("git init failed: %w", err)
		}
	}
	return nil
}

// ModTidy runs `go mod tidy` in outputPath. On failure the output directory is
// removed so the user is not left with a broken project.
func ModTidy(outputPath string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = outputPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		_ = os.RemoveAll(outputPath)
		return fmt.Errorf("go mod tidy failed (generated files removed): %w\n%s", err, output)
	}
	return nil
}

// GitInit runs `git init` in outputPath. It is a no-op when outputPath is
// already inside an existing git repository, so nested repos are never created.
func GitInit(outputPath string) error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not found in PATH: %w", err)
	}
	// Skip when already inside a git repository.
	check := exec.Command("git", "-C", outputPath, "rev-parse", "--is-inside-work-tree")
	if check.Run() == nil {
		return nil
	}
	cmd := exec.Command("git", "init")
	cmd.Dir = outputPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w\n%s", err, output)
	}
	return nil
}
