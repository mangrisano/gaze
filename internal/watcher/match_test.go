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
