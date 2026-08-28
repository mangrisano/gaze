package watcher

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

func TestRunOnceStreamsOutputOnSuccess(t *testing.T) {
	var out bytes.Buffer
	err := runOnce(context.Background(), false, "", &out, &out, "echo", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "hello") {
		t.Fatalf("stdout = %q, want it to contain %q", out.String(), "hello")
	}
}

func TestRunOnceReturnsErrorOnFailure(t *testing.T) {
	err := runOnce(context.Background(), false, "", io.Discard, io.Discard, "false")
	if err == nil {
		t.Fatal("want an error for a command that exits non-zero, got nil")
	}
}

func TestRunOnceSetsGazeFileEnv(t *testing.T) {
	var out bytes.Buffer
	err := runOnce(context.Background(), false, "internal/watcher/match.go", &out, &out, "sh", "-c", `printf %s "$GAZE_FILE"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := out.String(); got != "internal/watcher/match.go" {
		t.Fatalf("GAZE_FILE = %q, want %q", got, "internal/watcher/match.go")
	}
}

func TestRunOnceOmitsGazeFileWhenNoPath(t *testing.T) {
	var out bytes.Buffer
	// With no changed path we don't set GAZE_FILE, so it stays unset.
	err := runOnce(context.Background(), false, "", &out, &out, "sh", "-c", `printf %s "${GAZE_FILE-unset}"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := out.String(); got != "unset" {
		t.Fatalf("GAZE_FILE = %q, want it unset", got)
	}
}

func TestClearScreen(t *testing.T) {
	var buf bytes.Buffer
	clearScreen(&buf)
	if got, want := buf.String(), "\033[H\033[2J\033[3J"; got != want {
		t.Fatalf("clearScreen wrote %q, want %q", got, want)
	}
}
