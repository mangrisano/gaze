//go:build !windows

package watcher

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestRunOnceRestartTerminatesGracefully checks that in restart mode a cancelled
// command receives SIGTERM (which a process can trap) rather than an untrappable
// SIGKILL.
func TestRunOnceRestartTerminatesGracefully(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "caught-term")
	ready := filepath.Join(dir, "ready")

	// Trap SIGTERM (record it, then exit); signal readiness once the trap is set.
	script := `trap 'touch "` + marker + `"; exit 0' TERM
touch "` + ready + `"
while :; do sleep 0.02; done`

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		_ = runOnce(ctx, true, "", io.Discard, io.Discard, "sh", "-c", script)
		close(done)
	}()

	waitForFile(t, ready, 3*time.Second) // wait until the trap is installed
	cancel()                             // -> SIGTERM to the process group

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("runOnce did not return after cancel")
	}

	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("SIGTERM was not delivered gracefully (no marker): %v", err)
	}
}

func waitForFile(t *testing.T, path string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", path)
}
