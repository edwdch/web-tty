package main

import (
	"log"

	"github.com/joho/godotenv"

	"simple-app/internal/config"
	"simple-app/internal/httpserver"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	if err := httpserver.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
