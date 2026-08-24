package main

import (
	"log"

	"github.com/joho/godotenv"

	"github.com/edwdch/web-tty/internal/config"
	"github.com/edwdch/web-tty/internal/httpserver"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	if err := httpserver.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
