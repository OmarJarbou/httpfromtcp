package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/OmarJarbou/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan // blocking until get a signal
	log.Println("Server gracefully stopped")
	// This is a common pattern in Go for gracefully shutting down a server.
	// Because server.Serve returns immediately (it handles requests in the
	// background in goroutines) if we exit main immediately, the server will
	// just stop. We want to wait for a signal (like CTRL+C) before we stop
	// the server.
}
