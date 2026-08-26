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
	err := runOnce(context.Background(), false, &out, &out, "echo", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "hello") {
		t.Fatalf("stdout = %q, want it to contain %q", out.String(), "hello")
	}
}

func TestRunOnceReturnsErrorOnFailure(t *testing.T) {
	err := runOnce(context.Background(), false, io.Discard, io.Discard, "false")
	if err == nil {
		t.Fatal("want an error for a command that exits non-zero, got nil")
	}
}

func TestClearScreen(t *testing.T) {
	var buf bytes.Buffer
	clearScreen(&buf)
	if got, want := buf.String(), "\033[H\033[2J\033[3J"; got != want {
		t.Fatalf("clearScreen wrote %q, want %q", got, want)
	}
}
