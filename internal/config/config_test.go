package config

import (
	"reflect"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("ADDR", "")
	t.Setenv("GIN_MODE", "")
	t.Setenv("SHELL", "")
	t.Setenv("SHELL_ARGS", "")
	t.Setenv("CWD", "")
	t.Setenv("WRITABLE", "")
	t.Setenv("ALLOW_ORIGIN", "")
	t.Setenv("MAX_SESSIONS", "")
	t.Setenv("SESSION_IDLE", "")

	// Unset so getenv fallbacks apply. t.Setenv("") still sets the key;
	// clear by setting then using a subprocess-free approach: Load uses LookupEnv
	// and empty values fall through. Empty WRITABLE falls to true.
	cfg := Load()
	if cfg.Addr != ":8080" {
		t.Fatalf("Addr = %q", cfg.Addr)
	}
	if cfg.GinMode != "debug" {
		t.Fatalf("GinMode = %q", cfg.GinMode)
	}
	if cfg.Shell != "/bin/bash" {
		t.Fatalf("Shell = %q", cfg.Shell)
	}
	if !reflect.DeepEqual(cfg.ShellArgs, []string{"-l"}) {
		t.Fatalf("ShellArgs = %#v", cfg.ShellArgs)
	}
	if cfg.Cwd != "" {
		t.Fatalf("Cwd = %q", cfg.Cwd)
	}
	if !cfg.Writable {
		t.Fatal("Writable")
	}
	if cfg.AllowOrigin != nil {
		t.Fatalf("AllowOrigin = %#v", cfg.AllowOrigin)
	}
	if cfg.MaxSessions != 50 {
		t.Fatalf("MaxSessions = %d", cfg.MaxSessions)
	}
	if cfg.SessionIdle != 168*time.Hour {
		t.Fatalf("SessionIdle = %s", cfg.SessionIdle)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("ADDR", ":9090")
	t.Setenv("GIN_MODE", "release")
	t.Setenv("SHELL", "/bin/sh")
	t.Setenv("SHELL_ARGS", "-c echo")
	t.Setenv("CWD", "/tmp")
	t.Setenv("WRITABLE", "false")
	t.Setenv("ALLOW_ORIGIN", "http://127.0.0.1:5173, https://term.example.com")
	t.Setenv("MAX_SESSIONS", "3")
	t.Setenv("SESSION_IDLE", "24h")

	cfg := Load()
	if cfg.Addr != ":9090" || cfg.GinMode != "release" {
		t.Fatalf("http cfg %+v", cfg)
	}
	if cfg.Shell != "/bin/sh" {
		t.Fatalf("Shell = %q", cfg.Shell)
	}
	if !reflect.DeepEqual(cfg.ShellArgs, []string{"-c", "echo"}) {
		t.Fatalf("ShellArgs = %#v", cfg.ShellArgs)
	}
	if cfg.Cwd != "/tmp" {
		t.Fatalf("Cwd = %q", cfg.Cwd)
	}
	if cfg.Writable {
		t.Fatal("Writable should be false")
	}
	want := []string{"http://127.0.0.1:5173", "https://term.example.com"}
	if !reflect.DeepEqual(cfg.AllowOrigin, want) {
		t.Fatalf("AllowOrigin = %#v", cfg.AllowOrigin)
	}
	if cfg.MaxSessions != 3 {
		t.Fatalf("MaxSessions = %d", cfg.MaxSessions)
	}
	if cfg.SessionIdle != 24*time.Hour {
		t.Fatalf("SessionIdle = %s", cfg.SessionIdle)
	}
}

func TestLoadSessionIdleZeroAndInvalid(t *testing.T) {
	t.Setenv("SESSION_IDLE", "0")
	cfg := Load()
	if cfg.SessionIdle != 0 {
		t.Fatalf("zero idle = %s", cfg.SessionIdle)
	}

	t.Setenv("SESSION_IDLE", "nope")
	cfg = Load()
	if cfg.SessionIdle != 168*time.Hour {
		t.Fatalf("invalid idle = %s", cfg.SessionIdle)
	}
}
