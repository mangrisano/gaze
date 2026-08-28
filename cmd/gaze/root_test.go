package main

import (
	"testing"
	"time"
)

func TestRootCmdFlagShorthands(t *testing.T) {
	cmd := newRootCmd(nil)
	want := map[string]string{
		"ext":      "e",
		"path":     "p",
		"ignore":   "i",
		"debounce": "d",
		"clear":    "c",
		"restart":  "r",
		"grace":    "k",
		"version":  "v",
	}
	for name, short := range want {
		fl := cmd.Flags().Lookup(name)
		if fl == nil {
			t.Errorf("flag --%s not registered", name)
			continue
		}
		if fl.Shorthand != short {
			t.Errorf("--%s shorthand = %q, want %q", name, fl.Shorthand, short)
		}
	}
	if cmd.Flags().Lookup("no-initial") == nil {
		t.Error("flag --no-initial not registered")
	}
}

func TestRootCmdParsesFlags(t *testing.T) {
	cmd := newRootCmd([]string{"echo", "hi"})
	args := []string{"-e", "go", "--path", ".", "-k", "10s", "--no-initial", "-c"}
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}
	if got, _ := cmd.Flags().GetDuration("grace"); got != 10*time.Second {
		t.Errorf("grace = %v, want 10s", got)
	}
	if got, _ := cmd.Flags().GetBool("no-initial"); !got {
		t.Error("no-initial not set")
	}
	if exts, _ := cmd.Flags().GetStringArray("ext"); len(exts) != 1 || exts[0] != "go" {
		t.Errorf("ext = %v, want [go]", exts)
	}
}
