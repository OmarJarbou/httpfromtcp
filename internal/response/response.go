package response

import (
	"io"
	"strconv"

	"github.com/OmarJarbou/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	OK           StatusCode = 200
	CLIENT_ERROR StatusCode = 400
	SERVER_ERROR StatusCode = 500
)

func WriteStatusLine(w io.Writer, status_code StatusCode) error {
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

	_, err := w.Write([]byte(status_line))
	return err
}

func GetDefaultHeaders(content_Len int) headers.Headers {
	headers := headers.Headers{}

	headers["content-length"] = strconv.Itoa(content_Len)
	headers["connection"] = "close"
	headers["content-type"] = "text/plain"

	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	headers_text := ""

	for key, value := range headers {
		headers_text += key + ": " + value + "\r\n"
	}
	headers_text += "\r\n"

	_, err := w.Write([]byte(headers_text))
	return err
}
