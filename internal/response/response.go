package response

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"strconv"

	"github.com/OmarJarbou/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	OK           StatusCode = 200
	CLIENT_ERROR StatusCode = 400
	SERVER_ERROR StatusCode = 500
)

type WriterState int

const (
	STATUS_LINE WriterState = iota
	HEADERS
	BODY
)

type Writer struct {
	Writer      io.Writer
	WriterState WriterState
}

func WriterStateString(ws WriterState) string {
	write_state_string := ""
	switch ws {
	case STATUS_LINE:
		write_state_string = "status line"
	case HEADERS:
		write_state_string = "headers"
	case BODY:
		write_state_string = "body"
	}

	return write_state_string
}

func (w *Writer) WriteStatusLine(status_code StatusCode) error {
	if w.WriterState != STATUS_LINE {
		return errors.New("cant write " + WriterStateString(STATUS_LINE) + " now, you should write: " + WriterStateString(w.WriterState))
	}
	status_line := "HTTP/1.1 " + strconv.Itoa(int(status_code)) + " "

	switch status_code {
	case OK:
		status_line += "OK"
	case CLIENT_ERROR:
		status_line += "Bad Request"
	case SERVER_ERROR:
		status_line += "Internal Server Error"
	}

	status_line += "\r\n"

	_, err := w.Writer.Write([]byte(status_line))
	if err == nil {
		w.WriterState = HEADERS
	}
	return err
}

func isMimeType(s string) bool {
	_, _, err := mime.ParseMediaType(s)
	return err == nil
}

func GetDefaultHeaders(content_len int, content_type string) (headers.Headers, error) {
	headers := headers.Headers{}

	headers["content-length"] = strconv.Itoa(content_len)
	headers["connection"] = "close"
	if !isMimeType(content_type) {
		return nil, errors.New("invalid content type (should be a mime type): " + content_type)
	}
	headers["content-type"] = content_type

	return headers, nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.WriterState != HEADERS {
		return errors.New("cant write " + WriterStateString(HEADERS) + " now, you should write: " + WriterStateString(w.WriterState))
	}

	headers_text := ""

	for key, value := range headers {
		headers_text += key + ": " + value + "\r\n"
	}
	headers_text += "\r\n"

	_, err := w.Writer.Write([]byte(headers_text))
	if err == nil {
		w.WriterState = BODY
	}
	return err
}

func (w *Writer) WriteBody(data []byte) (int, error) {
	if w.WriterState != BODY {
		return 0, errors.New("cant write " + WriterStateString(BODY) + " now, you should write: " + WriterStateString(w.WriterState))
	}

	n, err := w.Writer.Write(data)
	return n, err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	chunked_body := ""
	chunk_length_in_hex := fmt.Sprintf("%X", len(p))
	chunked_body += chunk_length_in_hex + "\r\n" + string(p) + "\r\n"

	n, err := w.Writer.Write([]byte(chunked_body))
	return n, err
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	chunked_body_done := "0\r\n\r\n"

	n, err := w.Writer.Write([]byte(chunked_body_done))
	return n, err
}
