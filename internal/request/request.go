package request

import (
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/OmarJarbou/httpfromtcp/internal/headers"
)

type State int

const (
	initialized State = iota
	parsing_headers
	parsing_body
	done
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	ParserState State
}

func (r *Request) Get(header_name string) (header_value string, found bool) {
	header_value, found = r.Headers[strings.ToLower(header_name)]
	return
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
		Headers:     headers.Headers{},
		Body:        []byte{},
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
				if req.ParserState == parsing_headers {
					return nil, errors.New("header field has no ending")
				} else if req.ParserState == parsing_body {
					content_length_string, _ := req.Get("Content-Length")
					content_length, err := strconv.Atoi(content_length_string)
					if err != nil {
						return nil, err
					}
					if len(req.Body) == content_length {
						req.ParserState = done
					} else if len(req.Body) < content_length {
						return nil, errors.New("body of the request \"" + string(req.Body) + "\" is shorter than reported Content-Length")
					} else if len(req.Body) > content_length {
						return nil, errors.New("body of the request \"" + string(req.Body) + "\" is longer than reported Content-Length")
					}
				}
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
	// Since a single chunk (or buffer) can contain data for multiple headers, we can’t assume that
	// only one header exists per read. Therefore, instead of calling header.parse() just once per
	// chunk and clearing the buffer from only that header, we loop to parse all complete headers
	// currently available in the buffer. After successfully parsing each header, we remove its data
	// from the buffer, ensuring that only the unparsed (incomplete) data remains

	// While parsing request line is actualy done only one time, so it won't be affected

	totalBytesParsed := 0
	if r.ParserState == parsing_headers {
		for !(r.ParserState == parsing_body || r.ParserState == done) {
			n, err := r.parseSingle(data[totalBytesParsed:])
			totalBytesParsed += n
			if err != nil {
				return totalBytesParsed, err
			}
			if n == 0 {
				break
			}
		}
	}
	if r.ParserState == parsing_body || r.ParserState == initialized {
		n, err := r.parseSingle(data[totalBytesParsed:])
		totalBytesParsed += n
		if err != nil {
			return totalBytesParsed, err
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.ParserState {
	case initialized:
		n, req_line, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if req_line != nil {
			r.RequestLine = *req_line
			r.ParserState = parsing_headers
		}
		return n, nil
	case parsing_headers:
		n, headers_done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if headers_done {
			if _, ok := r.Get("Content-Length"); !ok {
				r.ParserState = done
			} else {
				r.ParserState = parsing_body
			}
		}
		return n, nil
	case parsing_body:
		n := r.parseBody(data)
		return n, nil
	case done:
		return 0, errors.New("error: trying to read data in a done state")
	default:
		return 0, errors.New("error: unknown state")
	}
}

func (r *Request) parseBody(data []byte) int {
	r.Body = append(r.Body, data...)
	return len(data)
}
