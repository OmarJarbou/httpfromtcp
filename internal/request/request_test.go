package request

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLineParse(t *testing.T) {
	// Test: Good GET Request line
	req, err := RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, req)
	assert.Equal(t, req.RequestLine.Method, "GET")
	assert.Equal(t, req.RequestLine.RequestTarget, "/")
	assert.Equal(t, req.RequestLine.HttpVersion, "HTTP/1.1")

	// Test: Good GET Request line with path
	req, err = RequestFromReader(strings.NewReader("GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, req)
	assert.Equal(t, req.RequestLine.Method, "GET")
	assert.Equal(t, req.RequestLine.RequestTarget, "/coffee")
	assert.Equal(t, req.RequestLine.HttpVersion, "HTTP/1.1")

	// Test: Good POST Request with path
	req, err = RequestFromReader(strings.NewReader("POST /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, req)
	assert.Equal(t, req.RequestLine.Method, "POST")
	assert.Equal(t, req.RequestLine.RequestTarget, "/coffee")
	assert.Equal(t, req.RequestLine.HttpVersion, "HTTP/1.1")

	// Test: Invalid number of parts in request line 1
	_, err = RequestFromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Invalid number of parts in request line 2
	_, err = RequestFromReader(strings.NewReader(" /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Invalid number of parts in request line 3
	_, err = RequestFromReader(strings.NewReader("GET HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Invalid number of parts in request line 4
	_, err = RequestFromReader(strings.NewReader("GET /coffee\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Invalid number of parts in request line 5
	_, err = RequestFromReader(strings.NewReader("GET /coffee HTTP/1.1 Hello\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Wrong use of small characters in method
	_, err = RequestFromReader(strings.NewReader("Get /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Invalid version in Request line 1
	_, err = RequestFromReader(strings.NewReader("GET /coffee HTTP/3.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Invalid method
	_, err = RequestFromReader(strings.NewReader("HELLO /coffee HTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Invalid method (out of order) Request line
	_, err = RequestFromReader(strings.NewReader("/coffee GET HTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)
}
