package watcher

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// collectDirs returns dir and all of its subdirectories (recursively), skipping
// any directory whose path contains an entry in ignore (that whole subtree is
// left out) and any subdirectory that can't be read (e.g. a protected dir).
func collectDirs(dir string, ignore []string) ([]string, error) {
	var dirs []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// The root itself being unreadable is a real error (e.g. a bad -p).
			if path == dir {
				return err
			}
			// A subtree we can't read (e.g. a macOS-protected dir like ~/.Trash):
			// WalkDir already added it on the errorless call, so drop it, skip its
			// contents, and keep scanning the rest.
			if d != nil && d.IsDir() {
				if n := len(dirs); n > 0 && dirs[n-1] == path {
					dirs = dirs[:n-1]
				}
				return filepath.SkipDir
			}
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		for _, ig := range ignore {
			if strings.Contains(path, ig) {
				return filepath.SkipDir
			}
		}
		dirs = append(dirs, path)
		return nil
	})
	return dirs, err
}
