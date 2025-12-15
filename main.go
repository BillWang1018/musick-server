package main

import (
	"log"

	"musick-server/internal/app"

	"github.com/joho/godotenv"
)

const listenAddr = "0.0.0.0:5896"

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	server := app.New()
	if err := server.Run(listenAddr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
