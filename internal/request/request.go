package request

import (
	"errors"
	"io"
	"strings"
)

type State int

const (
	initialized State = iota
	second
	done
)

type Request struct {
	RequestLine RequestLine
	ParserState State
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

const BUFFER_SIZE int = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := Request{
		RequestLine: RequestLine{},
		ParserState: initialized,
	}
	buffer := make([]byte, BUFFER_SIZE)
	expand_index := 1
	bytes_read_count := 0
	bytes_parsed_count := 0

	for req.ParserState != done {
		n, err := reader.Read(buffer[bytes_read_count:])
		if err != nil {
			if err == io.EOF {
				req.ParserState = done
				continue
			} else {
				return nil, err
			}
		}
		bytes_read_count += n
		if bytes_read_count == len(buffer) {
			temp_buf := buffer
			buffer = make([]byte, BUFFER_SIZE<<(expand_index)) //BUFFER_SIZE*math.Pow(2, expand_index)
			expand_index++
			copy(buffer, temp_buf)
		}

		p, err := req.parse(buffer[:bytes_read_count])
		if err != nil {
			return nil, err
		}
		bytes_parsed_count += p
		if p > 0 {
			copy(buffer, buffer[p:]) //Remove the data that was parsed successfully buffer[0:p-1] from the buffer
			bytes_read_count -= p
		}
	}

	return &req, nil
}

func parseRequestLine(req_bytes []byte) (int, *RequestLine, error) {
	req_line := RequestLine{}

	req_string := string(req_bytes)
	crlf_index := strings.Index(req_string, "\r\n")
	if crlf_index == -1 {
		return 0, nil, nil
	}
	req_parts := strings.Split(req_string, "\r\n")

	req_line_parts := strings.Split(req_parts[0], " ")
	if len(req_line_parts) != 3 {
		return 0, nil, errors.New("request line must contain 3 fundamental parts: METHOD, RREQUEST TARGET, HTTP VERSION")
	}

	for _, char := range req_line_parts[0] {
		if string(char) < "A" || string(char) > "Z" {
			return 0, nil, errors.New("\"" + req_line_parts[0] + "\": " + "method in request line must only contain capital alphabetic characters")
		}
	}
	if _, ok := supportedMethods[req_line_parts[0]]; !ok {
		return 0, nil, errors.New("\"" + req_line_parts[0] + "\": " + "method in request line should be one of the following: GET, HEAD, POST, PUT, DELETE, CONNECT, OPTIONS, TRACE")
	}

	http_version_parts := strings.Split(req_line_parts[2], "/")
	if len(http_version_parts) > 2 || http_version_parts[0] != "HTTP" || http_version_parts[1] != "1.1" {
		return 0, nil, errors.New("http version in request line must be HTTP/1.1")
	}

	req_line.HttpVersion = http_version_parts[1]
	req_line.RequestTarget = req_line_parts[1]
	req_line.Method = req_line_parts[0]

	return len(req_parts[0]) + 2 /*for crlf*/, &req_line, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.ParserState == initialized {
		n, req_line, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if req_line != nil {
			r.RequestLine = *req_line
			r.ParserState = second
		}
		return n, nil
	} else if r.ParserState == second {
		return 0, nil
	} else if r.ParserState == done {
		return 0, errors.New("error: trying to read data in a done state")
	} else {
		return 0, errors.New("error: unknown state")
	}
}
