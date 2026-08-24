// Package watcher watches directories and re-runs a command when files change.
package watcher

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Config holds the resolved options for a watch session.
type Config struct {
	Paths    []string      // directories to watch (recursively)
	Exts     []string      // file extensions that trigger a run; empty = all
	Ignore   []string      // path substrings to ignore
	Debounce time.Duration // coalesce a burst of events into one run
	Command  []string      // the command to run, name first
}

// Run watches the configured paths and runs the command once at startup and then
// on every relevant change. It blocks until ctx is cancelled.
func Run(ctx context.Context, cfg Config) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer w.Close()

	for _, p := range cfg.Paths {
		dirs, err := collectDirs(p, cfg.Ignore)
		if err != nil {
			return fmt.Errorf("cannot scan %q: %w", p, err)
		}
		for _, d := range dirs {
			if err := w.Add(d); err != nil {
				return fmt.Errorf("cannot watch %q: %w", d, err)
			}
		}
	}

	// Pipeline: raw pings -> debounce -> runLoop -> the command.
	in := make(chan struct{})
	trigger := debounce(in, cfg.Debounce)
	go runLoop(trigger, func(runCtx context.Context) {
		runOnce(runCtx, os.Stdout, os.Stderr, cfg.Command[0], cfg.Command[1:]...)
	})

	fmt.Printf("watching %s  cmd: %s\n", strings.Join(cfg.Paths, ", "), strings.Join(cfg.Command, " "))
	in <- struct{}{} // initial run on startup

	for {
		select {
		case event := <-w.Events:
			if shouldRun(event.Name, cfg.Exts, cfg.Ignore) {
				in <- struct{}{}
			}
		case err := <-w.Errors:
			fmt.Fprintln(os.Stderr, "watch error:", err)
		case <-ctx.Done():
			close(in)
			return nil
		}
	}
}
