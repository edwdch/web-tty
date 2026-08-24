package session

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

func TestHubCreateListClose(t *testing.T) {
	hub := testHub(t, 2, 0)
	h, err := hub.Create(80, 24)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	list := hub.List()
	if len(list) != 1 || list[0].ID != h.ID() || list[0].PID <= 0 {
		t.Fatalf("list = %+v", list)
	}
	if !hub.Has(h.ID()) || hub.Full() {
		t.Fatal("has/full")
	}
	if err := hub.Close(h.ID()); err != nil {
		t.Fatalf("close: %v", err)
	}
	if len(hub.List()) != 0 || hub.Has(h.ID()) {
		t.Fatalf("still listed: %+v", hub.List())
	}
	if err := hub.Close(h.ID()); err != ErrNotFound {
		t.Fatalf("second close: %v", err)
	}
}

func TestHubMaxSessions(t *testing.T) {
	hub := testHub(t, 1, 0)
	a, err := hub.Create(80, 24)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	if !hub.Full() {
		t.Fatal("expected full")
	}
	if _, err := hub.Create(80, 24); err != ErrFull {
		t.Fatalf("second: %v", err)
	}
	if err := hub.Close(a.ID()); err != nil {
		t.Fatalf("close: %v", err)
	}
	if _, err := hub.Create(80, 24); err != nil {
		t.Fatalf("after close: %v", err)
	}
}

func TestHubReplayAfterDetach(t *testing.T) {
	hub := testHub(t, 2, 0)
	h, err := hub.Create(80, 24)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	cid, out, _, err := h.AddClient()
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, err := h.Write([]byte("hello\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	waitClientContains(t, out, []byte("hello"))
	h.RemoveClient(cid)

	if len(hub.List()) != 1 {
		t.Fatalf("session dropped on detach: %+v", hub.List())
	}

	_, _, replay, err := h.AddClient()
	if err != nil {
		t.Fatalf("reattach: %v", err)
	}
	if !bytes.Contains(replay, []byte("hello")) {
		t.Fatalf("replay = %q", replay)
	}
}

func TestHubFanoutShare(t *testing.T) {
	hub := testHub(t, 2, 0)
	h, err := hub.Create(80, 24)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_, a, _, err := h.AddClient()
	if err != nil {
		t.Fatalf("a: %v", err)
	}
	_, b, _, err := h.AddClient()
	if err != nil {
		t.Fatalf("b: %v", err)
	}
	if _, err := h.Write([]byte("shared\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	waitClientContains(t, a, []byte("shared"))
	waitClientContains(t, b, []byte("shared"))
	if n := hub.List()[0].Clients; n != 2 {
		t.Fatalf("clients = %d", n)
	}
}

func TestHubShellExitRemovesSession(t *testing.T) {
	hub := NewHub(2, NewFactory(Options{
		Shell:     "/bin/sh",
		ShellArgs: []string{"-c", "echo bye; exit 0"},
	}), 0)
	t.Cleanup(hub.Stop)
	if _, err := hub.Create(80, 24); err != nil {
		skipMissing(t, "/bin/sh", err)
		t.Fatalf("create: %v", err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if len(hub.List()) == 0 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("session still listed: %+v", hub.List())
}

func TestHubIdleReapDetached(t *testing.T) {
	hub := testHub(t, 2, 80*time.Millisecond)
	h, err := hub.Create(80, 24)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	cid, _, _, err := h.AddClient()
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	h.RemoveClient(cid)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if len(hub.List()) == 0 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("idle session not reaped: %+v", hub.List())
}

func TestHubIdleSkipsAttached(t *testing.T) {
	hub := NewHub(2, NewFactory(Options{Shell: "/bin/sleep", ShellArgs: []string{"30"}}), 80*time.Millisecond)
	t.Cleanup(hub.Stop)
	h, err := hub.Create(80, 24)
	if err != nil {
		skipMissing(t, "/bin/sleep", err)
		t.Fatalf("create: %v", err)
	}
	cid, _, _, err := h.AddClient()
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	defer h.RemoveClient(cid)

	time.Sleep(300 * time.Millisecond)
	if len(hub.List()) != 1 {
		t.Fatalf("attached session reaped: %+v", hub.List())
	}
}

func testHub(t *testing.T, max int, idle time.Duration) *Hub {
	t.Helper()
	hub := NewHub(max, NewFactory(Options{Shell: "/bin/cat"}), idle)
	t.Cleanup(hub.Stop)
	return hub
}

func skipMissing(t *testing.T, shell string, err error) {
	t.Helper()
	if _, statErr := os.Stat(shell); statErr != nil {
		t.Skipf("missing %s: %v", shell, err)
	}
}

func waitClientContains(t *testing.T, out <-chan []byte, needle []byte) []byte {
	t.Helper()
	deadline := time.After(5 * time.Second)
	var buf []byte
	for {
		select {
		case <-deadline:
			t.Fatalf("timeout, got %q want %q", buf, needle)
		case frame, ok := <-out:
			if !ok {
				t.Fatalf("client closed, got %q want %q", buf, needle)
			}
			if len(frame) > 0 && frame[0] == MsgOutput {
				buf = append(buf, frame[1:]...)
			} else {
				buf = append(buf, frame...)
			}
			if bytes.Contains(buf, needle) {
				return buf
			}
		}
	}
}

func TestHubWindowTitleFromOSC(t *testing.T) {
	hub := testHub(t, 2, 0)
	h, err := hub.Create(80, 24)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !strings.HasPrefix(h.Title(), "cat") {
		t.Fatalf("base title = %q", h.Title())
	}
	_, out, _, err := h.AddClient()
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, err := h.Write([]byte("\x1b]0;grok 40%\x07\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	got := waitTitleContains(t, out, "grok 40%")
	if !strings.Contains(got, " | ") {
		t.Fatalf("expected ttyd-style suffix: %q", got)
	}
	if h.Title() != got {
		t.Fatalf("Title() = %q frame = %q", h.Title(), got)
	}

	if _, err := h.Write([]byte("\x1b]0;\x07\n")); err != nil {
		t.Fatalf("clear: %v", err)
	}
	cleared := waitTitleContains(t, out, "cat")
	if strings.Contains(cleared, "grok") {
		t.Fatalf("title not cleared: %q", cleared)
	}
}

func waitTitleContains(t *testing.T, out <-chan []byte, needle string) string {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatalf("timeout waiting for title %q", needle)
		case frame, ok := <-out:
			if !ok {
				t.Fatalf("client closed waiting for title %q", needle)
			}
			if len(frame) > 0 && frame[0] == MsgTitle && strings.Contains(string(frame[1:]), needle) {
				return string(frame[1:])
			}
		}
	}
}

func TestHubGetMissing(t *testing.T) {
	hub := testHub(t, 1, 0)
	if _, err := hub.Get("missing"); err != ErrNotFound {
		t.Fatalf("err = %v", err)
	}
}
