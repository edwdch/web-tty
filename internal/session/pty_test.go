package session

import (
	"bytes"
	"io"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestWithTermEnvOverrides(t *testing.T) {
	got := withTermEnv([]string{"PATH=/bin", "TERM=dumb", "COLORTERM=no", "HOME=/tmp"})
	want := []string{"PATH=/bin", "HOME=/tmp", "TERM=xterm-256color", "COLORTERM=truecolor"}
	if len(got) != len(want) {
		t.Fatalf("len=%d want %d (%q)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %q", got)
		}
	}
}

func TestSessionCatEcho(t *testing.T) {
	s := startTest(t, Options{Shell: "/bin/cat"}, 80, 24)
	defer s.Close()

	if _, err := s.Write([]byte("hello\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	got := readDeadline(t, s, 5*time.Second)
	if !bytes.Contains(got, []byte("hello")) {
		t.Fatalf("got %q", got)
	}
}

func TestSessionSttySizeAndResize(t *testing.T) {
	s := startTest(t, Options{
		Shell:     "/bin/sh",
		ShellArgs: []string{"-c", "stty size; echo READY; read x; stty size"},
	}, 80, 24)
	defer s.Close()

	first := readUntil(t, s, []byte("READY"), 5*time.Second)
	if !bytes.Contains(first, []byte("24 80")) {
		t.Fatalf("initial size = %q, want 24 80", first)
	}
	if err := s.Resize(50, 10); err != nil {
		t.Fatalf("resize: %v", err)
	}
	if _, err := s.Write([]byte("\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	second := readUntil(t, s, []byte("10 50"), 5*time.Second)
	if !bytes.Contains(second, []byte("10 50")) {
		t.Fatalf("resized = %q, want 10 50", second)
	}
}

func TestSessionCloseKillsProcess(t *testing.T) {
	s := startTest(t, Options{Shell: "/bin/sleep", ShellArgs: []string{"30"}}, 80, 24)
	pid := s.PID()
	if pid <= 0 {
		t.Fatal("missing pid")
	}
	if err := s.Close(); err != nil && err != io.EOF {
		// Wait can return a signal error; that's fine as long as the process is gone.
		t.Logf("close: %v", err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for {
		err := syscall.Kill(pid, 0)
		if err != nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("pid %d still alive", pid)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestTwoSessionsIndependentPIDs(t *testing.T) {
	a := startTest(t, Options{Shell: "/bin/cat"}, 80, 24)
	defer a.Close()
	b := startTest(t, Options{Shell: "/bin/cat"}, 80, 24)
	defer b.Close()
	if a.PID() == 0 || a.PID() == b.PID() {
		t.Fatalf("pids %d %d", a.PID(), b.PID())
	}
}

func TestHubMaxSessions(t *testing.T) {
	hub := NewHub(1, NewFactory(Options{Shell: "/bin/cat"}))
	if !hub.TryAcquire() {
		t.Fatal("first acquire")
	}
	if hub.TryAcquire() {
		t.Fatal("second acquire should fail")
	}
	hub.Release()
	if !hub.TryAcquire() {
		t.Fatal("acquire after release")
	}
}

func startTest(t *testing.T, opts Options, cols, rows uint16) *Session {
	t.Helper()
	s, err := Start(opts, cols, rows)
	if err != nil {
		if _, statErr := os.Stat(opts.Shell); statErr != nil {
			t.Skipf("missing %s: %v", opts.Shell, err)
		}
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func readDeadline(t *testing.T, r io.Reader, timeout time.Duration) []byte {
	t.Helper()
	return readUntil(t, r, nil, timeout)
}

func readUntil(t *testing.T, r io.Reader, needle []byte, timeout time.Duration) []byte {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var buf []byte
	tmp := make([]byte, 256)
	for {
		remain := time.Until(deadline)
		if remain <= 0 {
			t.Fatalf("timeout, got %q want %q", buf, needle)
		}
		type result struct {
			n   int
			err error
		}
		ch := make(chan result, 1)
		go func() {
			n, err := r.Read(tmp)
			ch <- result{n: n, err: err}
		}()
		var got result
		select {
		case got = <-ch:
		case <-time.After(remain):
			t.Fatalf("timeout, got %q want %q", buf, needle)
		}
		if got.n > 0 {
			buf = append(buf, tmp[:got.n]...)
			if len(needle) == 0 || bytes.Contains(buf, needle) {
				return buf
			}
		}
		if got.err != nil {
			if len(needle) == 0 && len(buf) > 0 {
				return buf
			}
			t.Fatalf("read: %v, buf=%q", got.err, buf)
		}
	}
}
