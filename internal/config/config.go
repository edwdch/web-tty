package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Addr        string
	GinMode     string
	Shell       string
	ShellArgs   []string
	Cwd         string
	Writable    bool
	AllowOrigin []string
	MaxSessions int
}

func Load() Config {
	return Config{
		Addr:        getenv("ADDR", ":8080"),
		GinMode:     getenv("GIN_MODE", "debug"),
		Shell:       getenv("SHELL", "/bin/bash"),
		ShellArgs:   strings.Fields(getenv("SHELL_ARGS", "-l")),
		Cwd:         getenv("CWD", ""),
		Writable:    getenvBool("WRITABLE", true),
		AllowOrigin: splitComma(getenv("ALLOW_ORIGIN", "")),
		MaxSessions: getenvInt("MAX_SESSIONS", 8),
	}
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getenvBool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func getenvInt(key string, fallback int) int {
	v := getenv(key, "")
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func splitComma(v string) []string {
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
