package server

import (
	"bytes"
	"io"
	"log"

	"github.com/OmarJarbou/httpfromtcp/internal/request"
	"github.com/OmarJarbou/httpfromtcp/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(io.Writer, *request.Request) *HandlerError

func (he *HandlerError) handlerResponseWriter(w io.Writer, buffer *bytes.Buffer) {
	err := response.WriteStatusLine(w, he.StatusCode)
	if err != nil {
		log.Fatal(err)
		return
	}
	headers := response.GetDefaultHeaders(len(buffer.Bytes()))
	err = response.WriteHeaders(w, headers)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = response.WriteBody(w, *buffer)
	if err != nil {
		log.Fatal(err)
		return
	}
}
