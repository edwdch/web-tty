package session

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
)

const killWait = 2 * time.Second

type Options struct {
	Shell     string
	ShellArgs []string
	Cwd       string
}

type Conn interface {
	io.ReadWriteCloser
	Resize(columns, rows uint16) error
	Pause()
	Resume()
}

type Session struct {
	cmd     *exec.Cmd
	pty     *os.File
	pauseMu sync.Mutex
	paused  bool
	unpause chan struct{}
	closed  chan struct{}

	closeOnce sync.Once
	closeErr  error
}

func Start(opts Options, columns, rows uint16) (*Session, error) {
	if opts.Shell == "" {
		opts.Shell = "/bin/bash"
	}
	cmd := exec.Command(opts.Shell, opts.ShellArgs...)
	if opts.Cwd != "" {
		cmd.Dir = opts.Cwd
	}
	cmd.Env = withTermEnv(os.Environ())

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: columns, Rows: rows})
	if err != nil {
		return nil, fmt.Errorf("start pty: %w", err)
	}

	return &Session{
		cmd:     cmd,
		pty:     ptmx,
		unpause: make(chan struct{}, 1),
		closed:  make(chan struct{}),
	}, nil
}

func (s *Session) PID() int {
	if s == nil || s.cmd == nil || s.cmd.Process == nil {
		return 0
	}
	return s.cmd.Process.Pid
}

func (s *Session) Read(p []byte) (int, error) {
	for {
		select {
		case <-s.closed:
			return 0, io.EOF
		default:
		}

		s.pauseMu.Lock()
		paused := s.paused
		s.pauseMu.Unlock()
		if !paused {
			return s.pty.Read(p)
		}

		select {
		case <-s.closed:
			return 0, io.EOF
		case <-s.unpause:
		}
	}
}

func (s *Session) Write(p []byte) (int, error) {
	return s.pty.Write(p)
}

func (s *Session) Resize(columns, rows uint16) error {
	sz := Size{Columns: columns, Rows: rows}
	if err := sz.Validate(); err != nil {
		return err
	}
	return pty.Setsize(s.pty, &pty.Winsize{Cols: columns, Rows: rows})
}

func (s *Session) Pause() {
	s.pauseMu.Lock()
	s.paused = true
	s.pauseMu.Unlock()
}

func (s *Session) Resume() {
	s.pauseMu.Lock()
	if !s.paused {
		s.pauseMu.Unlock()
		return
	}
	s.paused = false
	s.pauseMu.Unlock()
	select {
	case s.unpause <- struct{}{}:
	default:
	}
}

func (s *Session) Close() error {
	s.closeOnce.Do(func() {
		close(s.closed)
		if s.cmd.Process != nil {
			_ = s.cmd.Process.Signal(syscall.SIGHUP)
		}
		_ = s.pty.Close()

		done := make(chan error, 1)
		go func() {
			done <- s.cmd.Wait()
		}()
		select {
		case s.closeErr = <-done:
		case <-time.After(killWait):
			if s.cmd.Process != nil {
				_ = s.cmd.Process.Kill()
			}
			s.closeErr = <-done
		}
	})
	return s.closeErr
}

func withTermEnv(base []string) []string {
	out := make([]string, 0, len(base)+2)
	for _, e := range base {
		if strings.HasPrefix(e, "TERM=") || strings.HasPrefix(e, "COLORTERM=") {
			continue
		}
		out = append(out, e)
	}
	return append(out, "TERM=xterm-256color", "COLORTERM=truecolor")
}
