package watcher

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// collectDirs returns dir and all of its subdirectories (recursively), skipping
// any directory whose path contains an entry in ignore (that whole subtree is
// left out).
func collectDirs(dir string, ignore []string) ([]string, error) {
	var dirs []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
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
