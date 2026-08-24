// Command gaze re-runs a command whenever a watched file changes.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/mangrisano/gaze/internal/watcher"
)

// multiFlag collects a repeatable flag (e.g. -e go -e tmpl) into a slice.
type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func main() {
	own, command := splitArgs(os.Args[1:])

	fs := flag.NewFlagSet("gaze", flag.ExitOnError)
	var exts, paths, ignore multiFlag
	fs.Var(&exts, "e", "file extension to watch, repeatable (e.g. -e go)")
	fs.Var(&paths, "p", "directory to watch, repeatable (default .)")
	fs.Var(&ignore, "i", "path substring to ignore, repeatable (e.g. -i vendor)")
	delay := fs.Duration("d", 200*time.Millisecond, "debounce window")
	fs.Parse(own)

	if len(command) == 0 {
		fmt.Fprintln(os.Stderr, "usage: gaze [flags] -- <command> [args...]")
		os.Exit(2)
	}
	if len(paths) == 0 {
		paths = multiFlag{"."}
	}

	// A cancelled context on Ctrl-C flows into Run, which shuts the pipeline down.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg := watcher.Config{
		Paths:    paths,
		Exts:     exts,
		Ignore:   ignore,
		Debounce: *delay,
		Command:  command,
	}
	if err := watcher.Run(ctx, cfg); err != nil {
		log.Fatal(err)
	}
}
