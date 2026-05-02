// Copyright 2026 Iyad
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	Write(p []byte) (int, error)

	// WriteHeader sends the HTTP response header with the provided status code.
	// It can only be called once per response.
	WriteHeader(statusCode int)
}

type response struct {
	writer        *bufio.Writer
	header        Header
	statusCode    int
	headerWritten bool
	body          []byte
}

func (r *response) Header() Header {
	return r.header
}

func (r *response) Write(p []byte) (int, error) {
	r.WriteHeader(StatusOK)
	r.body = append(r.body, p...)
	return len(p), nil
}

func (r *response) WriteHeader(statusCode int) {
	if r.headerWritten {
		return
	}

	r.statusCode = statusCode
	r.headerWritten = true
}

func (r *response) flush() error {
	r.header.Set("Content-Length", strconv.Itoa(len(r.body)))

	r.writer.Write([]byte("HTTP/1.1 " + strconv.Itoa(r.statusCode) + " " + StatusText(r.statusCode)))
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
	r.writer.Write(r.body)
	return r.writer.Flush()
}
