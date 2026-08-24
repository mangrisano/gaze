package watcher

import (
	"testing"

	"github.com/fsnotify/fsnotify"
)

func TestShouldRun(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		op     fsnotify.Op
		exts   []string
		ignore []string
		want   bool
	}{
		{"no filter watches everything", "a/b/crev.go", fsnotify.Write, nil, nil, true},
		{"matching extension", "crev.go", fsnotify.Write, []string{"go"}, nil, true},
		{"non-matching extension", "notes.txt", fsnotify.Write, []string{"go"}, nil, false},
		{"one of several extensions", "page.tmpl", fsnotify.Write, []string{"go", "tmpl"}, nil, true},
		{"ignore beats a matching extension", "vendor/pkg/x.go", fsnotify.Write, []string{"go"}, []string{"vendor"}, false},
		{"ignore a dotdir with no ext filter", ".git/index", fsnotify.Write, nil, []string{".git"}, false},
		{"extension is a suffix, not a substring", "gopher.txt", fsnotify.Write, []string{"go"}, nil, false},
		{"chmod-only is ignored even with no filter", "main.go", fsnotify.Chmod, nil, nil, false},
		{"write plus chmod still runs", "main.go", fsnotify.Write | fsnotify.Chmod, []string{"go"}, nil, true},
		{"create triggers a run", "new.go", fsnotify.Create, []string{"go"}, nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldRun(tc.path, tc.op, tc.exts, tc.ignore); got != tc.want {
				t.Fatalf("shouldRun(%q, op=%v, exts=%v, ignore=%v) = %v, want %v",
					tc.path, tc.op, tc.exts, tc.ignore, got, tc.want)
			}
		})
	}
}

func TestWantsRun(t *testing.T) {
	files := map[string]bool{"notes/todo.txt": true}
	treeDirs := map[string]bool{"src": true}
	cases := []struct {
		name     string
		path     string
		op       fsnotify.Op
		exts     []string
		files    map[string]bool
		treeDirs map[string]bool
		want     bool
	}{
		{"exact file target fires", "notes/todo.txt", fsnotify.Write, nil, files, treeDirs, true},
		{"exact file ignores chmod-only", "notes/todo.txt", fsnotify.Chmod, nil, files, treeDirs, false},
		{"sibling in a file's dir is ignored", "notes/other.txt", fsnotify.Write, nil, files, treeDirs, false},
		{"file in a watched tree fires via extension", "src/main.go", fsnotify.Write, []string{"go"}, files, treeDirs, true},
		{"file in a watched tree with wrong extension", "src/readme.md", fsnotify.Write, []string{"go"}, files, treeDirs, false},
		{"pure dir mode delegates to shouldRun", "src/main.go", fsnotify.Write, []string{"go"}, nil, treeDirs, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := wantsRun(tc.path, tc.op, tc.exts, nil, tc.files, tc.treeDirs); got != tc.want {
				t.Fatalf("wantsRun(%q, op=%v) = %v, want %v", tc.path, tc.op, got, tc.want)
			}
		})
	}
}
