package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/OmarJarbou/httpfromtcp/internal/request"
	"github.com/OmarJarbou/httpfromtcp/internal/response"
	"github.com/OmarJarbou/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
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

func handler(w io.Writer, r *request.Request) *server.HandlerError {
	handler_err := server.HandlerError{}
	switch r.RequestLine.RequestTarget {
	case "/yourproblem":
		handler_err.StatusCode = response.CLIENT_ERROR
		handler_err.Message = "Your problem is not my problem\n"
	case "/myproblem":
		handler_err.StatusCode = response.SERVER_ERROR
		handler_err.Message = "Woopsie, my bad\n"
	default:
		handler_err.StatusCode = response.OK
		handler_err.Message = "All good, frfr\n"
	}
	_, err := w.Write([]byte(handler_err.Message))
	if err != nil {
		handler_err.StatusCode = response.SERVER_ERROR
		handler_err.Message = "Error while writing body to the buffer"
	}
	return &handler_err
}
