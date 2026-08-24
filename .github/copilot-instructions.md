# gaze — project context

A file-watcher CLI: re-runs a command whenever watched files change (like `entr`/`watchexec`).
Go learning project, **mentor mode**: the owner writes the code; the assistant scaffolds, guides, and reviews, and does NOT hand full solutions unless explicitly asked ("fallo tu"). Replies in Italian; code/comments in English.

## Build / test / run
- Test: `go test ./...` (add `-race` for the concurrency parts)
- Build: `go build -o gaze ./cmd/gaze` • Install: `go install ./cmd/gaze` (lands in `$(go env GOPATH)/bin`, must be on PATH)
- Use: `gaze [flags] -- <command> [args...]`, e.g. `gaze -e go -- go test ./...`
- Flags: `-e ext` (repeatable), `-p path` (repeatable, default `.`), `-i substr` (repeatable), `-d dur` (debounce, default 200ms). Everything after `--` is the command.

## Architecture (the pipeline)
Standard Go layout: `cmd/gaze` (package main) parses flags into a `watcher.Config` and calls `watcher.Run`; all logic lives in `internal/watcher` (public surface = `Config` + `Run`).
`fsnotify events → shouldRun (filter) → in chan → debounce → runLoop → runOnce → the command`
- `internal/watcher/match.go` — `shouldRun(path, exts, ignore)`: which file changes matter (ignore substring beats all; empty exts = all; else suffix `.ext`)
- `internal/watcher/debounce.go` — `debounce(in, d)`: coalesce a burst of events into one signal (`select` + `time.After` + nil-channel disarm)
- `internal/watcher/runner.go` — `runOnce(ctx, stdout, stderr, name, args...)`: `exec.CommandContext` + `cmd.Run`
- `internal/watcher/runloop.go` — `runLoop(trigger, run func(ctx))`: cancel the previous run when a new one starts (`context.WithCancel`)
- `internal/watcher/collectdirs.go` — `collectDirs(dir, ignore)`: `filepath.WalkDir` → dir + all subdirs (recursive watch), `SkipDir` on ignored
- `internal/watcher/watcher.go` — `Config` + `Run`: fsnotify setup (Add every dir from `collectDirs`), pipeline wiring, the initial ping, the `select` loop, ctx-driven shutdown
- `cmd/gaze/splitargs.go` — `splitArgs(args)`: split at the first `--` into own flags vs the command
- `cmd/gaze/main.go` — flag parsing (`multiFlag`), builds `watcher.Config`, calls `watcher.Run`, Ctrl-C via `signal.NotifyContext`

## Done
- `shouldRun`, `debounce`, `runOnce`, `runLoop`, `splitArgs`, `collectDirs` — all with tests, green (`go test ./...`, `-race` clean)
- **Recursive watching**: `collectDirs` walks the tree; `watcher.Run` Adds every subdir (verified live: editing a file in a subpackage triggers a run)
- Ctrl-C graceful shutdown: `signal.NotifyContext` → ctx cancel closes `in`, which dominoes down the pipeline (debounce returns → trigger closes → runLoop ends → cancels the running command)
- **Initial run on startup**: one ping on `in` before the select loop (verified live)
- **Standard layout**: split into `cmd/gaze` (package main) + `internal/watcher` (package watcher)
- **README** written; git repo created (github.com/mangrisano/gaze, identity gmail LOCAL)
- **Renamed watch → gaze**: the binary `watch` clashed with the Unix `watch(1)` command; module path is now `github.com/mangrisano/gaze`. Added LICENSE (MIT) + CI (.github/workflows/ci.yml)

## TODO / next
- Ignore `Chmod`-only events (less noise; on macOS `touch` can emit extra events → occasional double run)
- `-r` restart mode for long-running servers (kill + relaunch; `runLoop` already cancels)
- `-c` clear-screen before each run

## Notes / gotchas
- fsnotify is non-recursive by itself → hence `collectDirs`.
- Module path is `github.com/mangrisano/gaze`; internal import is `github.com/mangrisano/gaze/internal/watcher`. Install with `go install ./cmd/gaze` (root has no `package main`).
- Dep: `github.com/fsnotify/fsnotify`.
