// Package goversion provides a helper for deriving a Go version string from the
// running toolchain.
package goversion

import (
	"runtime"
	"strings"
)

// Default returns the current toolchain version as "major.minor" (e.g. "1.26"),
// stripping the patch segment and any build suffix if present.
func Default() string {
	v := strings.TrimPrefix(runtime.Version(), "go")
	// "1.26.0" → "1.26"; "1.26" stays "1.26"
	if first := strings.Index(v, "."); first != -1 {
		if second := strings.Index(v[first+1:], "."); second != -1 {
			v = v[:first+1+second]
		}
	}
	return v
}