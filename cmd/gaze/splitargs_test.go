package main

import "testing"

func TestSplitArgs(t *testing.T) {
	cases := []struct {
		name        string
		args        []string
		wantOwn     []string
		wantCommand []string
	}{
		{"flags then command", []string{"-e", "go", "--", "go", "test", "./..."}, []string{"-e", "go"}, []string{"go", "test", "./..."}},
		{"no separator: all own", []string{"-e", "go"}, []string{"-e", "go"}, nil},
		{"separator first: all command", []string{"--", "echo", "hi"}, nil, []string{"echo", "hi"}},
		{"separator with nothing after", []string{"--"}, nil, nil},
		{"empty", nil, nil, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			own, command := splitArgs(tc.args)
			if !equalArgs(own, tc.wantOwn) {
				t.Errorf("own = %#v, want %#v", own, tc.wantOwn)
			}
			if !equalArgs(command, tc.wantCommand) {
				t.Errorf("command = %#v, want %#v", command, tc.wantCommand)
			}
		})
	}
}

// equalArgs compares two string slices, treating nil and empty as equal.
func equalArgs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
