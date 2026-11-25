package util

import "path/filepath"

// KavaHomePath returns the local kvtool node home relative to tests/e2e.
func KavaHomePath() string {
	return filepath.Join(
		"..", "..", "build", ".cache", "kvtool",
		"full_configs", "generated", "kava", "initstate", ".kava",
	)
}
