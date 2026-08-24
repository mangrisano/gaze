# watch — project context

A file-watcher CLI: re-runs a command whenever watched files change (like `entr`/`watchexec`).
Go learning project, **mentor mode**: the owner writes the code; the assistant scaffolds, guides, and reviews, and does NOT hand full solutions unless explicitly asked ("fallo tu"). Replies in Italian; code/comments in English.

## Build / test / run
- Test: `go test ./...` (add `-race` for the concurrency parts)
- Build: `go build -o watch .` • Install: `go install .` (lands in `$(go env GOPATH)/bin`, must be on PATH)
- Use: `watch [flags] -- <command> [args...]`, e.g. `watch -e go -- go test ./...`
- Flags: `-e ext` (repeatable), `-p path` (repeatable, default `.`), `-i substr` (repeatable), `-d dur` (debounce, default 200ms). Everything after `--` is the command.

## Architecture (the pipeline)
`fsnotify events → shouldRun (filter) → in chan → debounce → runLoop → runOnce → the command`
- `match.go` — `shouldRun(path, exts, ignore)`: which file changes matter (ignore substring beats all; empty exts = all; else suffix `.ext`)
- `debounce.go` — `debounce(in, d)`: coalesce a burst of events into one signal (`select` + `time.After` + nil-channel disarm)
- `runner.go` — `runOnce(ctx, stdout, stderr, name, args...)`: `exec.CommandContext` + `cmd.Run`
- `runloop.go` — `runLoop(trigger, run func(ctx))`: cancel the previous run when a new one starts (`context.WithCancel`)
- `splitargs.go` — `splitArgs(args)`: split at the first `--` into own flags vs the command
- `collectdirs.go` — `collectDirs(dir, ignore)`: `filepath.WalkDir` → dir + all subdirs (recursive watch), `SkipDir` on ignored
- `main.go` — flag parsing (`multiFlag`), fsnotify setup (Add every dir from `collectDirs`), the pipeline wiring, the `select` loop, Ctrl-C shutdown

## Done
- `shouldRun`, `debounce`, `runOnce`, `runLoop`, `splitArgs`, `collectDirs` — all with tests, green (`go test ./...`, `-race` clean)
- **Recursive watching**: `collectDirs` walks the tree; `main.go` Adds every subdir (verified live: editing a file in a subpackage triggers a run)
- Ctrl-C graceful shutdown: `close(in)` dominoes down the pipeline (debounce returns → trigger closes → runLoop ends → cancels the running command)

## TODO / next
- **Initial run on startup** (today it waits for the first change; send one ping on `in` before the loop) — the easy next win
- Ignore `Chmod`-only events (less noise; on macOS `touch` can emit extra events → occasional double run)
- `-r` restart mode for long-running servers (kill + relaunch; `runLoop` already cancels)
- `-c` clear-screen before each run
- A README
- (optional) make it a git repo — ask before any git; gitignore the `watch` binary artifact

## Notes / gotchas
- fsnotify is non-recursive by itself → hence `collectDirs`.
- No initial run yet → after launch it prints `watching …` and waits; you must change a file to see it react.
- Dep: `github.com/fsnotify/fsnotify`.
