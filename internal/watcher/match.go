package watcher

import (
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// shouldRun reports whether a change to path should trigger a run.
func shouldRun(path string, op fsnotify.Op, exts []string, ignore []string) bool {
	if op == fsnotify.Chmod {
		return false
	}
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

// wantsRun routes a filesystem event using the exact file targets and the
// directory trees being watched. An exact file target fires only on its own
// path; a directory tree defers to shouldRun (extension/ignore/chmod). When any
// file target is set, changes to unrelated siblings in a file's parent dir are
// ignored unless that dir is also a directory target.
func wantsRun(name string, op fsnotify.Op, exts, ignore []string, files, treeDirs map[string]bool) bool {
	if files[filepath.Clean(name)] {
		return op != fsnotify.Chmod
	}
	if len(files) > 0 && !treeDirs[filepath.Clean(filepath.Dir(name))] {
		return false
	}
	return shouldRun(name, op, exts, ignore)
}
