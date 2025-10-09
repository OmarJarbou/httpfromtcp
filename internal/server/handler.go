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

func (hr *HandlerResponse) GetHeaders() headers.Headers {
	if hr.headers == nil {
		hr.headers = headers.Headers{}
	}
	return hr.headers
}

func (hr *HandlerResponse) ClearHeaders() {
	if hr.headers == nil {
		return
	}
	for key := range hr.headers {
		delete(hr.headers, key)
	}
}

type Handler func(response.Writer, *request.Request)

func (hr *HandlerResponse) HandlerResponseWriter(w response.Writer) {
	err := w.WriteStatusLine(hr.StatusCode)
	if err != nil {
		log.Println(err.Error())
		w.Close()
		return
	}
	if hr.headers["content-type"] == "" {
		hr.headers["content-type"] = "text/plain"
	}
	headers, err := response.GetDefaultHeaders(len(hr.Message), hr.headers["content-type"])
	if err != nil {
		log.Println(err.Error())
		w.Close()
		return
	}
	err = w.WriteHeaders(headers)
	if err != nil {
		log.Println(err.Error())
		w.Close()
		return
	}
	_, err = w.WriteBody([]byte(hr.Message))
	if err != nil {
		log.Println(err.Error())
		w.Close()
		return
	}
}

func (hr *HandlerResponse) HandlerErrorResponse(w response.Writer, StatusCode response.StatusCode, message string) {
	w.WriterState = response.STATUS_LINE
	hr.StatusCode = StatusCode
	hr.ClearHeaders()
	hr.SetHeader("Content-Type", "text/plain")
	hr.SetHeader("Connection", "close")
	hr.Message = message
	hr.HandlerResponseWriter(w)
}
