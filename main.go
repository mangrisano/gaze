// Command watch re-runs a command whenever a watched file changes.
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

	"github.com/fsnotify/fsnotify"
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

	fs := flag.NewFlagSet("watch", flag.ExitOnError)
	var exts, paths, ignore multiFlag
	fs.Var(&exts, "e", "file extension to watch, repeatable (e.g. -e go)")
	fs.Var(&paths, "p", "directory to watch, repeatable (default .)")
	fs.Var(&ignore, "i", "path substring to ignore, repeatable (e.g. -i vendor)")
	delay := fs.Duration("d", 200*time.Millisecond, "debounce window")
	fs.Parse(own)

	if len(command) == 0 {
		fmt.Fprintln(os.Stderr, "usage: watch [flags] -- <command> [args...]")
		os.Exit(2)
	}
	if len(paths) == 0 {
		paths = multiFlag{"."}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	for _, p := range paths {
		dirs, err := collectDirs(p, ignore)
		if err != nil {
			log.Fatalf("cannot scan %q: %v", p, err)
		}
		for _, d := range dirs {
			if err := watcher.Add(d); err != nil {
				log.Fatalf("cannot watch %q: %v", d, err)
			}
		}
	}

	// Pipeline: raw pings -> debounce -> runLoop -> the command.
	in := make(chan struct{})
	trigger := debounce(in, *delay)
	go runLoop(trigger, func(ctx context.Context) {
		runOnce(ctx, os.Stdout, os.Stderr, command[0], command[1:]...)
	})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	fmt.Printf("watching %s  cmd: %s\n", strings.Join(paths, ", "), strings.Join(command, " "))
	in <- struct{}{} // initial run on startup
	for {
		select {
		case event := <-watcher.Events:
			if shouldRun(event.Name, exts, ignore) {
				in <- struct{}{}
			}
		case err := <-watcher.Errors:
			fmt.Fprintln(os.Stderr, "watch error:", err)
		case <-stop:
			close(in)
			return
		}
	}
}
