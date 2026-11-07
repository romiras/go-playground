# file-watcher

Linux CLI that watches directory trees and prints one NDJSON event per line on **stdout**. Watcher errors and usage text go to **stderr** only.

Behavior and acceptance criteria are defined in [docs/spec.md](./docs/spec.md).

## Build

Requires Go **1.26.7+** and `GOEXPERIMENT=jsonv2` (the `go-filewatcher` dependency uses `encoding/json/v2`):

```bash
export GOEXPERIMENT=jsonv2
go build -o file-watcher .
```

From the playground workspace root, `go.work` can include `./file-watcher`.

## Usage

```text
file-watcher <root> [root...]
```

Each argument is a directory root. There are no flags. Missing roots print usage on stderr and exit 1.

```bash
GOEXPERIMENT=jsonv2 ./file-watcher ~/Documents
```

Symlink roots (for example `~/Documents`) are resolved with `filepath.EvalSymlinks` before watching so the tree under the target directory is actually monitored.

Only **`.md`** and **`.txt`** file events are emitted. Changes are debounced per path (100ms); `.gitignore` rules apply during the walk.

Stop with **Ctrl-C** or **SIGTERM**; buffered stdout is flushed and the process exits 0.

## Stdout (NDJSON)

One JSON object per line:

```json
{"timestamp":"2026-11-07T12:00:00Z","event":"modified","path":"/home/user/docs/notes.md","extension":".md","size_bytes":4096}
```

| `event` | Meaning |
| --- | --- |
| `created` | file created |
| `modified` | file written |
| `deleted` | file removed |
| `moved` | rename |

`touch` on an **existing** file usually produces no line (timestamp-only updates are not write/create events). Appending or rewriting content does.

## Tests

```bash
export GOEXPERIMENT=jsonv2
go test ./...
```

## Library

Built on [`github.com/larsartmann/go-filewatcher/v2`](https://github.com/LarsArtmann/go-filewatcher).
