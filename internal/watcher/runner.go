package watcher

import (
	"context"
	"io"
	"os/exec"
)

// runOnce runs name+args, streaming stdout/stderr to the given writers, and
// returns the command's exit error (nil on success).
func runOnce(ctx context.Context, stdout, stderr io.Writer, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	return err
}

// clearScreen writes the ANSI escape sequence that clears the terminal (screen
// and scrollback) so each run starts from a clean slate.
func clearScreen(w io.Writer) {
	io.WriteString(w, "\033[H\033[2J\033[3J")
}
