package config

import "os"

type Config struct {
	Addr    string
	GinMode string
}

func Load() Config {
	return Config{
		Addr:    getenv("ADDR", ":8080"),
		GinMode: getenv("GIN_MODE", "debug"),
	}
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
