//go:build windows

package watcher

import (
	"os/exec"
	"time"
)

// setupRestart is a no-op on Windows: Unix process groups and SIGTERM aren't
// available here, so the default context cancellation (Kill) is used instead.
func setupRestart(_ *exec.Cmd, _ time.Duration) {}
