// Command gaze re-runs a command whenever a watched file changes.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"time"

	"github.com/mangrisano/gaze/internal/watcher"
	"github.com/spf13/cobra"
)

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
	// splitArgs peels off the command after "--" so cobra only parses gaze's
	// own flags; the wrapped command never touches cobra's flag parser.
	own, command := splitArgs(os.Args[1:])
	root := newRootCmd(command)
	root.SetArgs(own)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// newRootCmd builds the gaze command. command is the argv after "--" (the
// program to run), already separated out by splitArgs.
func newRootCmd(command []string) *cobra.Command {
	var (
		exts     []string
		paths    []string
		ignore   []string
		clear    bool
		restart  bool
		noInit   bool
		debounce time.Duration
		grace    time.Duration
	)

	cmd := &cobra.Command{
		Use:   "gaze [flags] -- <command> [args...]",
		Short: "Re-run a command whenever a watched file changes",
		Long: "gaze watches files and re-runs a command on every change.\n" +
			"Everything after -- is the command to run.",
		Version:      resolveVersion(),
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			if len(command) == 0 {
				return fmt.Errorf("no command to run; usage: gaze [flags] -- <command> [args...]")
			}
			if len(paths) == 0 {
				paths = []string{"."}
			}
			// A cancelled context on Ctrl-C flows into Run, shutting the pipeline down.
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			return watcher.Run(ctx, watcher.Config{
				Paths:     paths,
				Exts:      exts,
				Ignore:    ignore,
				Debounce:  debounce,
				Clear:     clear,
				Restart:   restart,
				NoInitial: noInit,
				Grace:     grace,
				Command:   command,
			})
		},
	}

	f := cmd.Flags()
	f.StringArrayVarP(&exts, "ext", "e", nil, "file extension to watch, repeatable (e.g. -e go)")
	f.StringArrayVarP(&paths, "path", "p", nil, "file or directory to watch, repeatable (default .)")
	f.StringArrayVarP(&ignore, "ignore", "i", nil, "path substring to ignore, repeatable (e.g. -i vendor)")
	f.DurationVarP(&debounce, "debounce", "d", 200*time.Millisecond, "debounce window")
	f.BoolVarP(&clear, "clear", "c", false, "clear the terminal before each run")
	f.BoolVarP(&restart, "restart", "r", false, "restart mode: SIGTERM + process-group kill for long-running commands")
	f.DurationVarP(&grace, "grace", "k", 5*time.Second, "grace period before force-killing on restart (with -r)")
	f.BoolVar(&noInit, "no-initial", false, "skip the initial run on startup")
	f.BoolP("version", "v", false, "print version and exit")

	cmd.SetVersionTemplate("gaze {{.Version}}\n")

	// -e/-i/-d/-k take non-path values; -p keeps the default file completion.
	for _, name := range []string{"ext", "ignore", "debounce", "grace"} {
		_ = cmd.RegisterFlagCompletionFunc(name, cobra.NoFileCompletions)
	}

	return cmd
}
