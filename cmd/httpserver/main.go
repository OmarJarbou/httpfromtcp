package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
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

func handler(w response.Writer, r *request.Request) {
	handler_response := server.HandlerResponse{}
	switch r.RequestLine.RequestTarget {
	case "/yourproblem":
		handler_response.StatusCode = response.CLIENT_ERROR
		handler_response.SetHeader("Content-Type", "text/html")
		handler_response.Message = "Your request honestly kinda sucked."
	case "/myproblem":
		handler_response.StatusCode = response.SERVER_ERROR
		handler_response.SetHeader("Content-Type", "text/html")
		handler_response.Message = "Okay, you know what? This one is on me."
	default:
		handler_response.StatusCode = response.OK
		handler_response.SetHeader("Content-Type", "text/html")
		handler_response.Message = "Your request was an absolute banger."
	}
	handler_response.Message = htmlResponseFormat(handler_response.StatusCode, handler_response.Message)
	handler_response.HandlerResponseWriter(w)
}

func htmlResponseFormat(status_code response.StatusCode, message string) string {
	title := strconv.Itoa(int(status_code)) + " "
	status := ""
	switch status_code {
	case response.OK:
		status = "OK"
	case response.CLIENT_ERROR:
		status = "Bad Request"
	case response.SERVER_ERROR:
		status = "Internal Server Error"
	}
	title += status

	if status == "OK" {
		status = "Success!"
	}

	html_response := fmt.Sprintf("<html>\r\n\t<head>\r\n\t\t<title>%s</title>\r\n\t</head>\r\n\t<body>\r\n\t\t<h1>%s</h1>\r\n\t\t<p>%s</p>\r\n\t</body>\r\n</html>\r\n", title, status, message)
	return html_response
}
