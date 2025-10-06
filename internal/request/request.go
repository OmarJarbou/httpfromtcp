package request

import (
	"errors"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var supportedMethods map[string]struct{} = map[string]struct{}{
	"GET":     {},
	"HEAD":    {},
	"POST":    {},
	"PUT":     {},
	"DELETE":  {},
	"CONNECT": {},
	"OPTIONS": {},
	"TRACE":   {},
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := Request{}
	req_bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	req_string := string(req_bytes)
	req_parts := strings.Split(req_string, "\r\n")
	req_line, err := parseRequestLine(req_parts[0])
	if err != nil {
		return nil, err
	}

	req.RequestLine = req_line
	return &req, nil
}

func parseRequestLine(req_line_string string) (RequestLine, error) {
	req_line := RequestLine{}

	req_line_parts := strings.Split(req_line_string, " ")
	if len(req_line_parts) != 3 {
		return RequestLine{}, errors.New("request line must contain 3 fundamental parts: METHOD, RREQUEST TARGET, HTTP VERSION")
	}

	for _, char := range req_line_parts[0] {
		if string(char) < "A" || string(char) > "Z" {
			return req_line, errors.New("\"" + req_line_parts[0] + "\": " + "method in request line must only contain capital alphabetic characters")
		}
	}
	if _, ok := supportedMethods[req_line_parts[0]]; !ok {
		return req_line, errors.New("\"" + req_line_parts[0] + "\": " + "method in request line should be one of the following: GET, HEAD, POST, PUT, DELETE, CONNECT, OPTIONS, TRACE")
	}

	http_version_parts := strings.Split(req_line_parts[2], "/")
	if len(http_version_parts) > 2 || http_version_parts[0] != "HTTP" || http_version_parts[1] != "1.1" {
		return req_line, errors.New("http version in request line must be HTTP/1.1")
	}

	req_line.HttpVersion = req_line_parts[2]
	req_line.RequestTarget = req_line_parts[1]
	req_line.Method = req_line_parts[0]

	return req_line, nil
}
