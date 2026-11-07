package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	filewatcher "github.com/larsartmann/go-filewatcher/v2"
)

func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == "-cli" {
		os.Exit(run(os.Args[2:], os.Stdout, os.Stderr))
	}
	os.Exit(m.Run())
}

func TestRecordFrom(t *testing.T) {
	ts := time.Date(2026, 11, 7, 14, 0, 0, 0, time.FixedZone("EEST", 2*60*60))
	rec, ok := recordFrom(filewatcher.Event{
		Path:      "/home/user/docs/notes.md",
		Op:        filewatcher.Write,
		Timestamp: ts,
		Size:      4096,
	})
	if !ok {
		t.Fatal("expected mapped record")
	}
	got, err := json.Marshal(rec)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"timestamp":"2026-11-07T12:00:00Z","event":"modified","path":"/home/user/docs/notes.md","extension":".md","size_bytes":4096}`
	if string(got) != want {
		t.Fatalf("json = %s, want %s", got, want)
	}

	cases := []struct {
		op   filewatcher.Op
		name string
	}{
		{filewatcher.Create, "created"},
		{filewatcher.Write, "modified"},
		{filewatcher.Remove, "deleted"},
		{filewatcher.Rename, "moved"},
	}
	for _, tc := range cases {
		rec, ok := recordFrom(filewatcher.Event{Op: tc.op, Path: "a.txt", Timestamp: ts})
		if !ok || rec.Event != tc.name {
			t.Fatalf("op %v: got %q ok=%v", tc.op, rec.Event, ok)
		}
	}

	if _, ok := recordFrom(filewatcher.Event{Op: 0, Path: "a.txt"}); ok {
		t.Fatal("unknown op should be skipped")
	}
}

func TestNoArgs(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-cli")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected exit 1")
	}
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 {
		t.Fatalf("exit: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.String() != "usage: file-watcher <root> [root...]\n" {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestWatchTree(t *testing.T) {
	root := t.TempDir()
	notes := filepath.Join(root, "notes.md")
	if err := os.WriteFile(notes, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	ignoredDir := filepath.Join(root, "ignored")
	if err := os.Mkdir(ignoredDir, 0o755); err != nil {
		t.Fatal(err)
	}
	secret := filepath.Join(ignoredDir, "secret.md")
	if err := os.WriteFile(secret, []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("ignored\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	proc := startCLI(t, root)
	defer func() {
		proc.stop(t)
		if stderr := proc.stderr.String(); stderr != "" {
			t.Errorf("stderr = %q", stderr)
		}
	}()

	probe := filepath.Join(root, "probe.txt")
	if !proc.writeUntil(t, probe, nil, 3*time.Second, func(rec Record) bool {
		return rec.Path == probe && rec.Extension == ".txt"
	}) {
		t.Fatal("watcher did not report the probe file")
	}

	for i := 0; i < 5; i++ {
		if err := os.WriteFile(notes, []byte("burst"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	proc.expect(t, 3*time.Second, func(rec Record) bool {
		return rec.Path == notes && rec.Event == "modified" && rec.Extension == ".md" && rec.SizeBytes > 0
	})
	proc.expectQuiet(t, 500*time.Millisecond)

	txt := filepath.Join(root, "notes.txt")
	f, err := os.OpenFile(txt, os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	proc.expect(t, 3*time.Second, func(rec Record) bool {
		return rec.Path == txt && rec.Event == "created" && rec.Extension == ".txt"
	})
	if err := os.Remove(txt); err != nil {
		t.Fatal(err)
	}
	proc.expect(t, 3*time.Second, func(rec Record) bool {
		return rec.Path == txt && rec.Event == "deleted"
	})

	if err := os.WriteFile(filepath.Join(root, "skip.go"), []byte("package p\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	proc.expectQuiet(t, 500*time.Millisecond)

	sub := filepath.Join(root, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(sub, "note.md")
	if !proc.writeUntil(t, nested, []byte("n"), 3*time.Second, func(rec Record) bool {
		return rec.Path == nested && rec.Extension == ".md"
	}) {
		t.Fatal("nested .md change was not printed")
	}

	if err := os.WriteFile(secret, []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	proc.expectQuiet(t, 500*time.Millisecond)

	if err := os.WriteFile(notes, []byte("after-ignore"), 0o644); err != nil {
		t.Fatal(err)
	}
	proc.expect(t, 3*time.Second, func(rec Record) bool {
		return rec.Path == notes && rec.Event == "modified"
	})
}

func TestSymlinkRoot(t *testing.T) {
	real := t.TempDir()
	link := filepath.Join(t.TempDir(), "docs")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}

	proc := startCLI(t, link)
	defer proc.stop(t)

	path := filepath.Join(real, "1.txt")
	if !proc.writeUntil(t, path, []byte("x"), 3*time.Second, func(rec Record) bool {
		return rec.Path == path && rec.Extension == ".txt"
	}) {
		t.Fatal("symlink root produced no event")
	}
}

func TestInterruptFlushes(t *testing.T) {
	root := t.TempDir()
	proc := startCLI(t, root)

	path := filepath.Join(root, "notes.md")
	var rec Record
	if !proc.writeUntil(t, path, []byte("x"), 3*time.Second, func(got Record) bool {
		if got.Path != path {
			return false
		}
		rec = got
		return true
	}) {
		t.Fatal("watcher did not report notes.md")
	}
	if rec.Event != "created" && rec.Event != "modified" {
		t.Fatalf("event = %q", rec.Event)
	}

	if err := proc.cmd.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatal(err)
	}
	waitErr := proc.wait(t)
	if waitErr != nil {
		t.Fatalf("exit: %v\nstderr: %s", waitErr, proc.stderr.String())
	}
	rest := proc.drain()
	for _, line := range rest {
		var got Record
		if err := json.Unmarshal([]byte(line), &got); err != nil {
			t.Fatalf("trailing stdout %q: %v", line, err)
		}
	}
}

type cliProc struct {
	cmd    *exec.Cmd
	lines  chan string
	stderr *lockedBuf
	done   chan error
}

func startCLI(t *testing.T, root string) *cliProc {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-cli", root)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	stderr := &lockedBuf{}
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	lines := make(chan string, 32)
	go func() {
		sc := bufio.NewScanner(stdout)
		for sc.Scan() {
			lines <- sc.Text()
		}
		close(lines)
	}()
	p := &cliProc{cmd: cmd, lines: lines, stderr: stderr, done: make(chan error, 1)}
	go func() { p.done <- cmd.Wait() }()
	return p
}

func (p *cliProc) expect(t *testing.T, timeout time.Duration, pred func(Record) bool) Record {
	t.Helper()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case line, ok := <-p.lines:
			if !ok {
				t.Fatalf("stdout closed, stderr=%q", p.stderr.String())
			}
			var rec Record
			if err := json.Unmarshal([]byte(line), &rec); err != nil {
				t.Fatalf("stdout %q: %v", line, err)
			}
			if pred(rec) {
				return rec
			}
			t.Fatalf("unexpected event %s", line)
		case <-timer.C:
			t.Fatalf("timed out waiting for event, stderr=%q", p.stderr.String())
		}
	}
}

func (p *cliProc) writeUntil(t *testing.T, path string, body []byte, timeout time.Duration, pred func(Record) bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := os.WriteFile(path, body, 0o644); err != nil {
			t.Fatal(err)
		}
		timer := time.NewTimer(200 * time.Millisecond)
		select {
		case line, ok := <-p.lines:
			timer.Stop()
			if !ok {
				t.Fatalf("stdout closed, stderr=%q", p.stderr.String())
			}
			var rec Record
			if err := json.Unmarshal([]byte(line), &rec); err != nil {
				t.Fatalf("stdout %q: %v", line, err)
			}
			if pred(rec) {
				p.swallowPath(t, rec.Path, 400*time.Millisecond)
				return true
			}
			t.Fatalf("unexpected event %s", line)
		case <-timer.C:
		}
	}
	return false
}

func (p *cliProc) swallowPath(t *testing.T, path string, d time.Duration) {
	t.Helper()
	timer := time.NewTimer(d)
	defer timer.Stop()
	for {
		select {
		case line, ok := <-p.lines:
			if !ok {
				return
			}
			var rec Record
			if err := json.Unmarshal([]byte(line), &rec); err != nil {
				t.Fatalf("stdout %q: %v", line, err)
			}
			if rec.Path != path {
				t.Fatalf("unexpected event %s", line)
			}
		case <-timer.C:
			return
		}
	}
}

func (p *cliProc) expectQuiet(t *testing.T, d time.Duration) {
	t.Helper()
	select {
	case line, ok := <-p.lines:
		if ok {
			t.Fatalf("unexpected stdout %s\nstderr=%q", line, p.stderr.String())
		}
		t.Fatalf("stdout closed, stderr=%q", p.stderr.String())
	case <-time.After(d):
	}
}

func (p *cliProc) wait(t *testing.T) error {
	t.Helper()
	select {
	case err := <-p.done:
		return err
	case <-time.After(5 * time.Second):
		_ = p.cmd.Process.Kill()
		<-p.done
		t.Fatal("timeout waiting for exit")
	}
	return nil
}

func (p *cliProc) drain() []string {
	var lines []string
	for line := range p.lines {
		lines = append(lines, line)
	}
	return lines
}

type lockedBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *lockedBuf) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *lockedBuf) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}

func (p *cliProc) stop(t *testing.T) {
	t.Helper()
	if p.cmd.Process == nil {
		return
	}
	_ = p.cmd.Process.Signal(syscall.SIGINT)
	if err := p.wait(t); err != nil {
		t.Fatalf("stop: %v\nstderr: %s", err, p.stderr.String())
	}
}
