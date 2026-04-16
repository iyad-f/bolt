package bolt

import (
	"bufio"
	"strconv"
)

var (
	crlf       = []byte("\r\n")
	colonSpace = []byte(": ")
)

// ResponseWriter is used by handlers to construct an HTTP response.
type ResponseWriter interface {
	// Header returns the response header map that will be sent by WriteHeader.
	Header() Header

	// Write writes the data to the connection as part of an HTTP response.
	// If WriteHeader has not been called, it calls WriteHeader(200) before writing the data.
	Write(bytes []byte) (int, error)

	// WriteHeader sends the HTTP response header with the provided status code.
	// It can only be called once per response.
	WriteHeader(statusCode int)
}

type response struct {
	writer        *bufio.Writer
	header        Header
	statusCode    int
	headerWritten bool
}

func (r *response) Header() Header {
	return r.header
}

func (r *response) Write(bytes []byte) (int, error) {
	if !r.headerWritten {
		r.WriteHeader(StatusOK)
	}
	return r.writer.Write(bytes)
}

func (r *response) WriteHeader(statusCode int) {
	if r.headerWritten {
		return
	}

	r.statusCode = statusCode
	r.writer.Write([]byte("HTTP/1.1 " + strconv.Itoa(statusCode) + " " + StatusText(statusCode)))
	r.writer.Write(crlf)

	for key, values := range r.header {
		for _, value := range values {
			r.writer.Write([]byte(key))
			r.writer.Write(colonSpace)
			r.writer.Write([]byte(value))
			r.writer.Write(crlf)
		}
	}

	r.writer.Write(crlf)
	r.headerWritten = true
}
