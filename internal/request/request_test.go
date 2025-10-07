package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func TestRequestLineParse(t *testing.T) {
	// Test: Good GET Request line 1
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	req, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, req)
	assert.Equal(t, req.RequestLine.Method, "GET")
	assert.Equal(t, req.RequestLine.RequestTarget, "/")
	assert.Equal(t, req.RequestLine.HttpVersion, "1.1")

	// Test: Good GET Request line 2
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 40,
	}
	req, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, req)
	assert.Equal(t, req.RequestLine.Method, "GET")
	assert.Equal(t, req.RequestLine.RequestTarget, "/")
	assert.Equal(t, req.RequestLine.HttpVersion, "1.1")

	// Test: Good GET Request line 3
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 79,
	}
	req, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, req)
	assert.Equal(t, req.RequestLine.Method, "GET")
	assert.Equal(t, req.RequestLine.RequestTarget, "/")
	assert.Equal(t, req.RequestLine.HttpVersion, "1.1")

	// Test: Good GET Request line with path
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	req, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, req)
	assert.Equal(t, req.RequestLine.Method, "GET")
	assert.Equal(t, req.RequestLine.RequestTarget, "/coffee")
	assert.Equal(t, req.RequestLine.HttpVersion, "1.1")

	// Test: Good POST Request with path
	reader = &chunkReader{
		data:            "POST /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	req, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, req)
	assert.Equal(t, req.RequestLine.Method, "POST")
	assert.Equal(t, req.RequestLine.RequestTarget, "/coffee")
	assert.Equal(t, req.RequestLine.HttpVersion, "1.1")

	// Test: Invalid number of parts in request line 1
	reader = &chunkReader{
		data:            "/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Invalid number of parts in request line 2
	reader = &chunkReader{
		data:            " /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Invalid number of parts in request line 3
	reader = &chunkReader{
		data:            "GET HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Invalid number of parts in request line 4
	reader = &chunkReader{
		data:            "GET /coffee\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Invalid number of parts in request line 5
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1 Hello\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Wrong use of small characters in method
	reader = &chunkReader{
		data:            "Get /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Invalid version in Request line 1
	reader = &chunkReader{
		data:            "GET /coffee HTTP/3.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Invalid method
	reader = &chunkReader{
		data:            "HELLO /coffee HTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Invalid method (out of order) Request line
	reader = &chunkReader{
		data:            "/coffee GET HTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)
}

func TestHeadersParse(t *testing.T) {
	// Test: Standard Headers
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers["host"])
	assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
	assert.Equal(t, "*/*", r.Headers["accept"])

	// Test: Standard Headers 2
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 14,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers["host"])
	assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
	assert.Equal(t, "*/*", r.Headers["accept"])

	// Test: Malformed Header
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Empty Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\n\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)

	// Test: Duplicate Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nHost: localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers["host"])

	// Test: Same Header with differnt values
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nHost: localhost:42070\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069, localhost:42070", r.Headers["host"])

	// Test: Missing End of Headers
	// reader = &chunkReader{
	// 	data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nHost: localhost:42070\r\n",
	// 	numBytesPerRead: 3,
	// }
	// r, err = RequestFromReader(reader)
	// require.Error(t, err)
	// require.NotNil(t, r)
	// assert.Equal(t, "localhost:42069, localhost:42070", r.Headers["host"])

	// Test: Empty header value
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: \r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "", r.Headers["host"])
}
