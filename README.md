# gaze

<div align="center">

[![CI](https://github.com/mangrisano/gaze/actions/workflows/ci.yml/badge.svg)](https://github.com/mangrisano/gaze/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/tag/mangrisano/gaze?sort=semver&label=release)](https://github.com/mangrisano/gaze/tags)
[![Downloads](https://img.shields.io/github/downloads/mangrisano/gaze/total?label=downloads)](https://github.com/mangrisano/gaze/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

</div>

A small file-watcher CLI: it re-runs a command whenever watched files change — like [`entr`](https://eradman.com/entrproject/) or [`watchexec`](https://github.com/watchexec/watchexec), but minimal and written in Go.

```sh
gaze -e go -- go test ./...
```

Runs your tests once immediately, then again on every save.

## Features

- **Runs on startup**, then on every relevant change (no need to touch a file to get the first run).
- **Recursive watching** — every subdirectory is watched, not just the top level.
- **Debouncing** — a burst of events (editor writes, `git checkout`, etc.) is coalesced into a single run.
- **Cancels the previous run** when a new change arrives, so runs never pile up.
- **Filtering** by extension and by ignored path substrings.
- **Tells the command which file changed** via the `GAZE_FILE` environment variable.
- **Graceful shutdown** on Ctrl-C (the running command is cancelled cleanly).

## Install

### Homebrew (macOS / Linux)

```sh
brew install mangrisano/gaze/gaze
```

### Go

```sh
go install github.com/mangrisano/gaze/cmd/gaze@latest
```

Or from a local clone:

```sh
go install ./cmd/gaze
```

The binary lands in `$(go env GOPATH)/bin`, which must be on your `PATH`.

### Pre-built binaries

No Go toolchain? Grab a ready-made binary for your platform (Linux, macOS and Windows, `amd64`/`arm64`) from the [latest release](https://github.com/mangrisano/gaze/releases/latest). For example, on macOS (arm64):

```sh
VERSION=0.1.2
curl -sL "https://github.com/mangrisano/gaze/releases/download/v${VERSION}/gaze_${VERSION}_darwin_arm64.tar.gz" | tar xz
./gaze --version
```

## Usage

```
gaze [flags] -- <command> [args...]
```

Everything after `--` is the command to run.

| Flag           | Repeatable | Default     | Description                                                                                                      |
| -------------- | :--------: | ----------- | ---------------------------------------------------------------------------------------------------------------- |
| `-e ext`       |    yes     | (all files) | Only react to files with this extension (e.g. `-e go`).                                                          |
| `-p path`      |    yes     | `.`         | File or directory to watch (dirs are recursive).                                                                 |
| `-i substr`    |    yes     | —           | Ignore any path containing this substring (e.g. `-i vendor`).                                                    |
| `-d dur`       |     no     | `200ms`     | Debounce window (any Go duration, e.g. `500ms`, `1s`).                                                           |
| `-c`           |     no     | off         | Clear the terminal before each run.                                                                              |
| `-r`           |     no     | off         | Restart mode for long-running commands: on change, SIGTERM the whole process group (graceful), then start fresh. |
| `-k dur`       |     no     | `5s`        | Grace period before force-killing on restart (with `-r`), e.g. `-k 10s`.                                         |
| `--no-initial` |     no     | off         | Skip the run on startup; only run when a file actually changes.                                                  |
| `--version`    |     no     | —           | Print the version and exit (also `-v`).                                                                          |

## Examples

Re-run tests on any Go change:

```sh
gaze -e go -- go test ./...
```

Watch two directories, ignore `vendor`, rebuild:

```sh
gaze -p ./cmd -p ./internal -i vendor -e go -- go build ./...
```

Longer debounce for noisy editors:

```sh
gaze -d 500ms -e go -- go vet ./...
```

Watch a single file — only that file triggers a run (its parent dir is watched, so editor "atomic saves" are caught too):

```sh
gaze -p notes.txt -- cat notes.txt
```

Restart a server on change — `-r` gracefully SIGTERMs the whole process group, so the child frees its port before the new instance starts:

```sh
gaze -r -e go -- go run ./cmd/server
```

Act on the file that changed — gaze exports its path as `GAZE_FILE` (unset on the startup run, so use a `:-` fallback):

```sh
# test only the package of the changed file
gaze -e go -- sh -c 'go test ./"$(dirname "${GAZE_FILE:-.}")"'

# lint just the file that changed
gaze -e py -- sh -c 'ruff check "${GAZE_FILE:-.}"'
```

The path is the one fsnotify reports, relative to the watched root (e.g. `internal/watcher/match.go`). When a burst of files changes at once, `GAZE_FILE` holds the last one.

## How it works

The core is a small pipeline of channels:

```
fsnotify events → shouldRun (filter) → debounce → runLoop → runOnce → your command
```

- **`shouldRun`** decides which file changes matter (ignore substring wins; empty `-e` means all files; otherwise the path must end in `.ext`).
- **`debounce`** collapses a burst of events into one signal using `select` + `time.After`, carrying the last changed path.
- **`runLoop`** starts the command for each signal and cancels the previous run (via `context`) when a new one arrives.
- **`runOnce`** runs the command with `exec.CommandContext`, wiring stdout/stderr through and exporting the changed file as `GAZE_FILE`.

fsnotify is not recursive on its own, so `collectDirs` walks the tree with `filepath.WalkDir` and every directory is added to the watcher. Directories created while gaze is running are detected and watched too.

## Project layout

```
cmd/gaze/          package main — flag parsing and wiring
internal/watcher/  package watcher — the pipeline (Config + Run)
```

`cmd/gaze` only parses flags into a `watcher.Config` and calls `watcher.Run`; all the logic lives in `internal/watcher`, whose only public surface is `Config` and `Run`.

## Development

```sh
go build ./...
go test ./...          # add -race for the concurrency parts
go vet ./...
```

To embed the version in the binary, inject it at build time:

```sh
go build -ldflags "-X main.version=$(git describe --tags)" -o gaze ./cmd/gaze
```

Requires Go 1.27+. The only dependency is [`github.com/fsnotify/fsnotify`](https://github.com/fsnotify/fsnotify).

## License

MIT — see [LICENSE](LICENSE).
