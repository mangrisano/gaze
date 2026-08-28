//go:build !windows

package watcher

import (
	"os/exec"
	"syscall"
	"time"
)

// setupRestart makes cmd run in its own process group so the whole tree can be
// terminated together. When cmd's context is cancelled, SIGTERM is sent to the
// group for a graceful shutdown; if the process has not exited after the grace
// period, os/exec force-kills it.
func setupRestart(cmd *exec.Cmd, grace time.Duration) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		// A negative pid targets the whole group (pgid == the leader's pid).
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	}
	cmd.WaitDelay = grace
}
