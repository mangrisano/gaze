package watcher

import "testing"

func TestShouldRun(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		exts   []string
		ignore []string
		want   bool
	}{
		{"no filter watches everything", "a/b/crev.go", nil, nil, true},
		{"matching extension", "crev.go", []string{"go"}, nil, true},
		{"non-matching extension", "notes.txt", []string{"go"}, nil, false},
		{"one of several extensions", "page.tmpl", []string{"go", "tmpl"}, nil, true},
		{"ignore beats a matching extension", "vendor/pkg/x.go", []string{"go"}, []string{"vendor"}, false},
		{"ignore a dotdir with no ext filter", ".git/index", nil, []string{".git"}, false},
		{"extension is a suffix, not a substring", "gopher.txt", []string{"go"}, nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldRun(tc.path, tc.exts, tc.ignore); got != tc.want {
				t.Fatalf("shouldRun(%q, exts=%v, ignore=%v) = %v, want %v",
					tc.path, tc.exts, tc.ignore, got, tc.want)
			}
		})
	}
}
