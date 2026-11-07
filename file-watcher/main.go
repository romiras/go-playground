package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	filewatcher "github.com/larsartmann/go-filewatcher/v2"
)

const outCap = 64

// Record is one NDJSON object on stdout.
type Record struct {
	Timestamp string `json:"timestamp"`
	Event     string `json:"event"`
	Path      string `json:"path"`
	Extension string `json:"extension"`
	SizeBytes int64  `json:"size_bytes"`
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	diag := newDiag(stderr)
	if len(args) == 0 {
		diag.printf("usage: file-watcher <root> [root...]\n")
		return 1
	}

	roots, err := resolveRoots(args)
	if err != nil {
		diag.printf("%v\n", err)
		return 1
	}

	watcher, err := filewatcher.New(roots,
		filewatcher.WithExtensions(".md", ".txt"),
		filewatcher.WithPerPathDebounce(100*time.Millisecond),
		filewatcher.WithGitignore(true),
		filewatcher.WithMiddleware(
			filewatcher.MiddlewareDeduplicate(200*time.Millisecond),
			filewatcher.MiddlewareRecovery(),
		),
		filewatcher.WithErrorHandler(diag.errorHandler),
	)
	if err != nil {
		diag.printf("%v\n", err)
		return 1
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	events, err := watcher.Watch(ctx)
	if err != nil {
		diag.printf("%v\n", err)
		_ = watcher.Close()
		return 1
	}
	defer func() { _ = watcher.Close() }()

	go func() {
		<-ctx.Done()
		_ = watcher.Close()
	}()

	out := make(chan Record, outCap)
	var reader sync.WaitGroup
	reader.Add(1)
	go func() {
		defer reader.Done()
		defer close(out)
		for ev := range events {
			rec, ok := recordFrom(ev)
			if !ok {
				continue
			}
			out <- rec
		}
	}()

	if err := writeRecords(out, stdout); err != nil {
		diag.printf("%v\n", err)
		reader.Wait()
		return 1
	}
	reader.Wait()
	return 0
}

// resolveRoots turns each CLI root into the directory that should be watched.
// filepath.WalkDir does not follow a symlink root, and the watcher then records
// that path without adding an fsnotify watch. EvalSymlinks makes a symlink to a
// directory, such as ~/Documents, watch the tree it names.
func resolveRoots(args []string) ([]string, error) {
	roots := make([]string, len(args))
	for i, arg := range args {
		resolved, err := filepath.EvalSymlinks(arg)
		if err != nil {
			return nil, err
		}
		roots[i] = resolved
	}
	return roots, nil
}

func recordFrom(ev filewatcher.Event) (Record, bool) {
	name, ok := eventName(ev.Op)
	if !ok {
		return Record{}, false
	}
	return Record{
		Timestamp: ev.Timestamp.UTC().Format(time.RFC3339),
		Event:     name,
		Path:      ev.Path,
		Extension: filepath.Ext(ev.Path),
		SizeBytes: ev.Size,
	}, true
}

func eventName(op filewatcher.Op) (string, bool) {
	switch op {
	case filewatcher.Create:
		return "created", true
	case filewatcher.Write:
		return "modified", true
	case filewatcher.Remove:
		return "deleted", true
	case filewatcher.Rename:
		return "moved", true
	default:
		return "", false
	}
}

// writeRecords is the only writer of stdout. It flushes whenever out is
// empty and again before returning.
func writeRecords(out <-chan Record, stdout io.Writer) error {
	bw := bufio.NewWriter(stdout)
	for {
		rec, ok := <-out
		if !ok {
			return bw.Flush()
		}
		if err := writeRecord(bw, rec); err != nil {
			return err
		}
		if err := flushDrained(bw, out); err != nil {
			return err
		}
	}
}

func flushDrained(bw *bufio.Writer, out <-chan Record) error {
	for {
		select {
		case rec, ok := <-out:
			if !ok {
				return bw.Flush()
			}
			if err := writeRecord(bw, rec); err != nil {
				return err
			}
		default:
			return bw.Flush()
		}
	}
}

func writeRecord(bw *bufio.Writer, rec Record) error {
	payload, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	if _, err := bw.Write(payload); err != nil {
		return err
	}
	return bw.WriteByte('\n')
}

// diag is the only writer of stderr.
type diag struct {
	mu sync.Mutex
	w  io.Writer
}

func newDiag(w io.Writer) *diag {
	return &diag{w: w}
}

func (d *diag) printf(format string, args ...any) {
	d.mu.Lock()
	defer d.mu.Unlock()
	fmt.Fprintf(d.w, format, args...)
}

func (d *diag) errorHandler(ctx filewatcher.ErrorContext, err error) {
	switch {
	case ctx.Operation != "" && ctx.Path != "":
		d.printf("%s: %s: %v\n", ctx.Operation, ctx.Path, err)
	case ctx.Operation != "":
		d.printf("%s: %v\n", ctx.Operation, err)
	default:
		d.printf("%v\n", err)
	}
}
