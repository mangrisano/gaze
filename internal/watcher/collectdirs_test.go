package watcher

import (
	"os"
	"path/filepath"
	"slices"
	"sort"
	"testing"
)

func TestCollectDirs(t *testing.T) {
	root := t.TempDir()

	// Build a tree: root/a, root/a/b, root/vendor, root/vendor/x
	for _, d := range []string{"a", filepath.Join("a", "b"), "vendor", filepath.Join("vendor", "x")} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// A file, to prove files are not returned (only directories).
	if err := os.WriteFile(filepath.Join(root, "a", "f.go"), []byte("package a"), 0o644); err != nil {
		t.Fatal(err)
	}

	dirs, err := collectDirs(root, []string{"vendor"})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{root, filepath.Join(root, "a"), filepath.Join(root, "a", "b")}
	sort.Strings(dirs)
	sort.Strings(want)
	if !slices.Equal(dirs, want) {
		t.Fatalf("collectDirs = %v, want %v", dirs, want)
	}
}
