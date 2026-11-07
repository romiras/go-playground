# Simplified file-watcher

A Linux CLI that watches directory trees and prints one NDJSON event per line on stdout. This is the reduced form of [spec-full.md](./spec-full.md): no systemd unit, no config file, no log file.

Library: [`github.com/larsartmann/go-filewatcher/v2`](https://github.com/LarsArtmann/go-filewatcher). Context7 indexes fsnotify (`/fsnotify/fsnotify`) and does not index this module. The options below follow that module’s README (v2) and the practical default in [Go-ecosystem.md](./Go-ecosystem.md).

## 1. Scope

In scope:

- Positional directory roots. No flags.
- Recursive watch and runtime registration of new subdirectories, inside go-filewatcher (`WithRecursive` defaults to true).
- `.md` and `.txt` only.
- Per-path debounce, gitignore filtering, dedup, and panic recovery from the practical default.
- NDJSON records on stdout through an application-owned buffered channel.
- Diagnostics on stderr only.

Out of scope:

- systemd unit, linger, journal.
- `--watch`, `--config`, `--log`, and any JSON config file.
- Append-only log writer (`O_APPEND` / mutex file writer).
- Startup abort when prospective watches would exceed 80% of `max_user_watches`. On budget exhaustion the library skips directories and keeps running.
- Polling, content hashing, metrics, self-heal, circuit breaker.

## 2. Build

Upstream requires Go 1.26.7 or later and `GOEXPERIMENT=jsonv2` (the module imports `encoding/json/v2`). The playground workspace file `go.work` says `go 1.26`; the module toolchain must still satisfy 1.26.7.

```bash
export GOEXPERIMENT=jsonv2
go get github.com/larsartmann/go-filewatcher/v2
```

## 3. CLI

```text
file-watcher <root> [root...]
```

Each argument is a directory root passed to `filewatcher.New`. Missing arguments: usage text on stderr, exit 1, no bytes on stdout.

## 4. Watcher options

```go
filewatcher.New(roots,
    filewatcher.WithExtensions(".md", ".txt"),
    filewatcher.WithPerPathDebounce(100*time.Millisecond),
    filewatcher.WithGitignore(true),
    filewatcher.WithMiddleware(
        filewatcher.MiddlewareDeduplicate(200*time.Millisecond),
        filewatcher.MiddlewareRecovery(),
    ),
    filewatcher.WithErrorHandler(stderrOnly),
)
```

`WithGitignore(true)` matches the library default; set it explicitly. `WithBuffer` stays at the library default of 64. `stderrOnly` replaces the default error handler, which logs to stderr, so library logging cannot also print events.

`Watch(ctx)` returns the event channel. `Chmod` is dropped by the library. Extension filtering applies to emitted events; directory recursion still watches new subdirectories.

## 5. Pipeline

```text
Watch(ctx) events
       │
       ▼
reader goroutine ── record ──► out chan (capacity 64)
                                      │
                                      ▼
                            writer goroutine
                            sole user of stdout
                            json line + bufio flush
```

- The reader is the only consumer of the `Watch` channel. It maps each event to the record in section 6 and sends it on `out`.
- `out` is `make(chan Record, 64)`. A full buffer blocks the sender. Records are not dropped.
- The writer is the only goroutine that touches stdout. It marshals one JSON object, writes a newline, and uses `bufio.Writer`. It flushes when the channel is drained and before exit.
- A separate diagnostics path is the only user of stderr (section 7).

## 6. Stdout record

One JSON object per line, no extra log prefix:

```json
{"timestamp":"2026-11-07T12:00:00Z","event":"modified","path":"/home/user/docs/notes.md","extension":".md","size_bytes":4096}
```

| Field | Source |
| --- | --- |
| `timestamp` | `Event.Timestamp` in UTC, `time.RFC3339` |
| `event` | Op map below |
| `path` | `Event.Path` |
| `extension` | `filepath.Ext(Event.Path)` |
| `size_bytes` | `Event.Size`; `0` when unavailable (including typical deletes) |

| `Event.Op` | `event` |
| --- | --- |
| `Create` | `created` |
| `Write` | `modified` |
| `Remove` | `deleted` |
| `Rename` | `moved` |

Stdout carries these lines and nothing else.

## 7. Stderr

Stderr carries usage text, `New` / `Watch` failures, and watcher errors. One diagnostics writer.

`WithErrorHandler` is the watcher-error path. Do not also print `watcher.Errors()` when that handler already consumes the same error.

Startup failure (`New` or `Watch` returns an error, or roots are missing): write the error to stderr, write nothing to stdout, exit 1.

## 8. Shutdown

On `SIGINT` or `SIGTERM`:

1. Close the watcher (cancel the watch context).
2. Let the reader finish, then close `out`.
3. Let the writer drain `out` and flush stdout.
4. Exit 0.

## 9. Acceptance

- `file-watcher` with no args exits 1; stderr explains usage; stdout is empty.
- `file-watcher /path/to/tree` watches that tree recursively.
- Saving `notes.md` yields one stdout line with `"event":"modified"` after the 100ms per-path quiet window. A burst of writes on that path collapses to one line.
- Creating `notes.txt` yields `"event":"created"`. Deleting it yields `"event":"deleted"`.
- A `.go` file change produces no stdout line.
- A new subdirectory created at runtime is watched; a later `.md` change inside it is printed.
- Paths ignored by `.gitignore` produce no stdout line.
- Watcher errors appear on stderr and never on stdout. Event lines never appear on stderr.
- After Ctrl-C, stdout is flushed and the process exits 0.
