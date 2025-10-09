package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/OmarJarbou/httpfromtcp/internal/request"
	"github.com/OmarJarbou/httpfromtcp/internal/response"
	"github.com/OmarJarbou/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, videoHandler)
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

func videoHandler(w response.Writer, r *request.Request) {
	handler_response := server.HandlerResponse{}
	if r.RequestLine.RequestTarget == "/video" {
		handler_response.StatusCode = response.OK

		handler_response.SetHeader("Content-Type", "video/mp4")
		handler_response.SetHeader("Connection", "close")

		err := w.WriteStatusLine(handler_response.StatusCode)
		if err != nil {
			log.Println("Error while writing status line: " + err.Error())
			w.Close()
			return
		}
		err = w.WriteHeaders(handler_response.GetHeaders())
		if err != nil {
			log.Println("Error while writing headers: " + err.Error())
			w.Close()
			return
		}

		video_data, err := os.ReadFile("assets/vim.mp4")
		if err != nil {
			log.Println("Error while reading video file: " + err.Error())
			w.Close()
			return
		}
		_, err = w.WriteBody(video_data)
		if err != nil {
			log.Println("Error while writing body: " + err.Error())
			w.Close()
			return
		}
	}
}

func proxyHandler(w response.Writer, r *request.Request) {
	handler_response := server.HandlerResponse{}
	if strings.HasPrefix(r.RequestLine.RequestTarget, "/httpbin") {
		client := &http.Client{}
		url := "https://httpbin.org" + strings.TrimPrefix(r.RequestLine.RequestTarget, "/httpbin")
		req, err := client.Get(url)
		if err != nil {
			handler_response.HandlerErrorResponse(w, response.CLIENT_ERROR, "Error while making request to \""+url+"\": "+err.Error())
			return
		} else {
			status, err := strconv.Atoi(strings.Split(req.Status, " ")[0])
			if err != nil {
				handler_response.HandlerErrorResponse(w, response.SERVER_ERROR, "Error while parsing status of response from \""+url+"\": "+err.Error())
				return
			}

			if status >= 200 && status <= 299 {
				handler_response.StatusCode = response.OK
			} else if status >= 400 && status <= 499 {
				handler_response.StatusCode = response.CLIENT_ERROR
			} else if status >= 500 && status <= 599 {
				handler_response.StatusCode = response.SERVER_ERROR
			} else {
				handler_response.StatusCode = response.OK
			}

			handler_response.SetHeader("Content-Type", "text/plain")
			handler_response.SetHeader("Connection", "close")
			handler_response.SetHeader("Transfer-Encoding", "chunked")
			handler_response.SetHeader("Trailer", "X-Content-SHA256, X-Content-Length")

			err = w.WriteStatusLine(handler_response.StatusCode)
			if err != nil {
				log.Println("Error while writing status line: " + err.Error())
				w.Close()
				return
			}
			err = w.WriteHeaders(handler_response.GetHeaders())
			if err != nil {
				log.Println("Error while writing headers: " + err.Error())
				w.Close()
				return
			}

			hasher := sha256.New()
			body := make([]byte, 1024)
			body_bytes := 0
			for {
				n, body_read_err := req.Body.Read(body)
				body_bytes += n
				fmt.Println(n)
				if body_read_err != nil && !(body_read_err == io.EOF && n > 0) {
					if body_read_err == io.EOF {
						break
					}
					log.Println("Error while parsing body of response from \"" + url + "\": " + body_read_err.Error())
					w.Close()
					return
				}

				hasher.Write(body[:n])
				_, err = w.WriteChunkedBody(body[:n])
				if err != nil {
					log.Println("Error while writing a body chunk: " + err.Error())
					w.Close()
					return
				}

				if body_read_err == io.EOF {
					break
				}
			}
			_, err = w.WriteChunkedBodyDone()
			if err != nil {
				log.Println("Error while writing body done chunk: " + err.Error())
				w.Close()
				return
			}

			hash := hasher.Sum(nil)
			handler_response.SetHeader("X-Content-SHA256", fmt.Sprintf("%x", hash))
			handler_response.SetHeader("X-Content-Length", strconv.Itoa(body_bytes))
			err = w.WriteTrailers(handler_response.GetHeaders())
			if err != nil {
				log.Println("Error while writing trailers: " + err.Error())
				w.Close()
				return
			}
		}
	}
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
