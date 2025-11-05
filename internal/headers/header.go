package headers

import (
	"errors"
	"regexp"
	"strings"
)

type Headers map[string]string

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	headers_string := string(data)

	crlf_first_occurrence := strings.Index(headers_string, "\r\n")
	if crlf_first_occurrence == -1 {
		return 0, false, nil
	}

	if crlf_first_occurrence == 0 {
		return 2, true, nil
	}

	headers_from_data := strings.Split(headers_string, "\r\n")
	first_colon_occurrence := strings.Index(headers_from_data[0], ":")
	if first_colon_occurrence == -1 {
		return 0, false, errors.New("a header/field-line should contain a \":\" to split field-name and field-value")
	}
	// ex: header = "  Host : localhost:42069  "
	header_name := headers_from_data[0][:first_colon_occurrence]    // = "  Host "
	header_value := headers_from_data[0][first_colon_occurrence+1:] // = " localhost:42069  "
	header_name = strings.TrimLeft(header_name, " ")                // = "Host "
	header_value = strings.Trim(header_value, " ")                  // = "localhost:42069"

	if len(header_name) < 1 {
		return 0, false, errors.New("header-name must be at least of length 1")
	}

	if strings.Contains(header_name, " ") {
		return 0, false, errors.New("the field-name in header/field-line must not contain whitespaces after it (i.e. before the colon)")
	}

	match, err := regexp.MatchString("^[A-Za-z0-9!#$%&'*+-.^_`|~]+$", header_name)
	if err != nil {
		return 0, false, err
	}
	if !match {
		return 0, false, errors.New("header-name can contain only: capital letters, small letters, digits, and special characters (!,#,$,%,&,',*,+,-,.,^,_,`,|,~)")
	}

	if value, ok := h[strings.ToLower(header_name)]; ok {
		if value != header_value {
			h[strings.ToLower(header_name)] += ", " + header_value
		}
	} else {
		h[strings.ToLower(header_name)] = header_value
	}
	consumed_bytes := len(headers_from_data[0]) + /*crlf; because we removed it on split*/ 2

	return consumed_bytes, false, nil
}
