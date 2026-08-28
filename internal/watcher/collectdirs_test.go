package watcher

import (
	"os"
	"path/filepath"
	"runtime"
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

func TestCollectDirsSkipsUnreadableDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod 0o000 does not restrict directory reads on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	root := t.TempDir()

	readable := filepath.Join(root, "readable")
	blocked := filepath.Join(root, "blocked")
	for _, d := range []string{readable, blocked} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chmod(blocked, 0o000); err != nil {
		t.Fatal(err)
	}
	// Restore perms so t.TempDir cleanup can remove the tree.
	t.Cleanup(func() { _ = os.Chmod(blocked, 0o755) })

	dirs, err := collectDirs(root, nil)
	if err != nil {
		t.Fatalf("collectDirs errored on an unreadable subdir: %v", err)
	}
	if !slices.Contains(dirs, readable) {
		t.Fatalf("collectDirs = %v, want it to include %q", dirs, readable)
	}
	if slices.Contains(dirs, blocked) {
		t.Fatalf("collectDirs = %v, should not include the unreadable %q", dirs, blocked)
	}
}
