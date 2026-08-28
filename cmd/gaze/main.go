// Command gaze re-runs a command whenever a watched file changes.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
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

// version is overwritten at build time via -ldflags "-X main.version=...".
var version = "dev"

// resolveVersion prefers an ldflags-injected version, then the module version
// embedded by `go install`, then the VCS commit of a local build.
func resolveVersion() string {
	if version != "dev" {
		return version
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return version
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" {
			rev := s.Value
			if len(rev) > 7 {
				rev = rev[:7]
			}
			return "dev (" + rev + ")"
		}
	}
	return version
}

func main() {
	own, command := splitArgs(os.Args[1:])

	fs := flag.NewFlagSet("gaze", flag.ExitOnError)
	var exts, paths, ignore multiFlag
	fs.Var(&exts, "e", "file extension to watch, repeatable (e.g. -e go)")
	fs.Var(&paths, "p", "file or directory to watch, repeatable (default .)")
	fs.Var(&ignore, "i", "path substring to ignore, repeatable (e.g. -i vendor)")
	clear := fs.Bool("c", false, "clear the terminal before watching")
	restart := fs.Bool("r", false, "restart mode: SIGTERM + process-group kill for long running commands")
	grace := fs.Duration("k", 5*time.Second, "grace period before force-killing on restart (with -r)")
	noInitial := fs.Bool("no-initial", false, "skip the initial run on startup")
	delay := fs.Duration("d", 200*time.Millisecond, "debounce window")
	showVersion := fs.Bool("v", false, "print version and exit")
	fs.BoolVar(showVersion, "version", false, "print version and exit")
	fs.Parse(own)

	if *showVersion {
		fmt.Println("gaze", resolveVersion())
		return
	}

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
		Paths:     paths,
		Exts:      exts,
		Ignore:    ignore,
		Debounce:  *delay,
		Clear:     *clear,
		Restart:   *restart,
		NoInitial: *noInitial,
		Grace:     *grace,
		Command:   command,
	}
	if err := watcher.Run(ctx, cfg); err != nil {
		log.Fatal(err)
	}
}
