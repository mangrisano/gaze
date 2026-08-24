package watcher

import "strings"

// shouldRun reports whether a change to path should trigger a run.
func shouldRun(path string, exts []string, ignore []string) bool {
	for _, val := range ignore {
		if strings.Contains(path, val) {
			return false
		}
	}
	if len(exts) == 0 {
		return true
	}
	for _, ext := range exts {
		if strings.HasSuffix(path, "."+ext) {
			return true
		}
	}
	return false
}
