package server

import (
	"log"
	"strings"

	"github.com/OmarJarbou/httpfromtcp/internal/headers"
	"github.com/OmarJarbou/httpfromtcp/internal/request"
	"github.com/OmarJarbou/httpfromtcp/internal/response"
)

type HandlerResponse struct {
	StatusCode response.StatusCode
	headers    headers.Headers
	Message    string
}

func (hr *HandlerResponse) SetHeader(key, value string) {
	if hr.headers == nil {
		hr.headers = headers.Headers{}
	}
	hr.headers[strings.ToLower(key)] = value
}

type Handler func(response.Writer, *request.Request)

func (hr *HandlerResponse) HandlerResponseWriter(w response.Writer) {
	err := w.WriteStatusLine(hr.StatusCode)
	if err != nil {
		log.Fatal(err)
		return
	}
	if hr.headers["content-type"] == "" {
		hr.headers["content-type"] = "text/plain"
	}
	headers, err := response.GetHeaders(len(hr.Message), hr.headers["content-type"])
	if err != nil {
		log.Fatal(err)
		return
	}
	err = w.WriteHeaders(headers)
	if err != nil {
		log.Fatal(err)
		return
	}
	_, err = w.WriteBody([]byte(hr.Message))
	if err != nil {
		log.Fatal(err)
		return
	}
}
