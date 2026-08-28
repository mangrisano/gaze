// Package watcher watches directories and re-runs a command when files change.
package watcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Config holds the resolved options for a watch session.
type Config struct {
	Paths     []string      // files or directories to watch (dirs recursively)
	Exts      []string      // file extensions that trigger a run; empty = all
	Clear     bool          // if true, clean the terminal and scrollback
	Restart   bool          // restart mode
	NoInitial bool          // skip the run on startup
	Ignore    []string      // path substrings to ignore
	Debounce  time.Duration // coalesce a burst of events into one run
	Grace     time.Duration // grace period before force-kill on restart
	Command   []string      // the command to run, name first
}

// Run watches the configured paths and runs the command once at startup and then
// on every relevant change. It blocks until ctx is cancelled.
func Run(ctx context.Context, cfg Config) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer w.Close()

	// Split targets into exact files and directory trees. A file is watched via
	// its parent dir (so editor atomic saves are still caught) but fires only on
	// its exact path; a directory is watched recursively and filtered by -e/-i.
	files := make(map[string]bool)
	treeDirs := make(map[string]bool)
	watch := func(d string) {
		if err := w.Add(d); err != nil {
			// Best-effort: skip a dir we can't watch (sockets, protected
			// paths, ...) rather than refusing to start at all.
			fmt.Fprintf(os.Stderr, "gaze: skipping %q: %v\n", d, err)
		}
	}
	// addDirs records each dir as a watched tree and starts watching it.
	addDirs := func(dirs []string) {
		for _, d := range dirs {
			treeDirs[filepath.Clean(d)] = true
			watch(d)
		}
	}

	for _, p := range cfg.Paths {
		info, err := os.Stat(p)
		if err != nil {
			return fmt.Errorf("cannot access %q: %w", p, err)
		}
		if info.IsDir() {
			dirs, err := collectDirs(p, cfg.Ignore)
			if err != nil {
				return fmt.Errorf("cannot scan %q: %w", p, err)
			}
			addDirs(dirs)
			continue
		}
		files[filepath.Clean(p)] = true
		watch(filepath.Dir(p))
	}

	// Pipeline: raw pings -> debounce -> runLoop -> the command.
	in := make(chan string)
	trigger := debounce(in, cfg.Debounce)
	go runLoop(trigger, func(runCtx context.Context, path string) {
		if cfg.Clear {
			clearScreen(os.Stdout)
		}
		runOnce(runCtx, cfg.Restart, cfg.Grace, path, os.Stdout, os.Stderr, cfg.Command[0], cfg.Command[1:]...)
	})

	fmt.Printf("watching %s  cmd: %s\n", strings.Join(cfg.Paths, ", "), strings.Join(cfg.Command, " "))
	if !cfg.NoInitial {
		in <- ""
	}

	for {
		select {
		case event := <-w.Events:
			// A new directory inside a watched tree isn't watched by fsnotify on
			// its own, so add its subtree as soon as it appears.
			if shouldWatchNewDir(event.Name, event.Op, treeDirs) {
				if dirs, err := collectDirs(event.Name, cfg.Ignore); err == nil {
					addDirs(dirs)
				}
			}
			if wantsRun(event.Name, event.Op, cfg.Exts, cfg.Ignore, files, treeDirs) {
				in <- event.Name
			}
		case err := <-w.Errors:
			fmt.Fprintln(os.Stderr, "watch error:", err)
		case <-ctx.Done():
			close(in)
			return nil
		}
	}
}
