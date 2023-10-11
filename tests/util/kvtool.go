package util

import (
	"path/filepath"
)

// KavaHomePath returns the OS-specific filepath for the kava home directory
// Assumes network is running with kvtool installed from the sub-repository in tests/e2e/kvtool
func KavaHomePath() string {
	return filepath.Join("kvtool", "full_configs", "generated", "kava", "initstate", ".kava")
}
