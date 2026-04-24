package bolt

import (
	"bufio"
	"io"
	"strings"
	"testing"
)

func TestReadRequest(t *testing.T) {
	t.Run("valid GET", func(t *testing.T) {
		br := bufio.NewReader(strings.NewReader("GET /users HTTP/1.1\r\nHost: localhost\r\n\r\n"))
		req, err := ReadRequest(br)
		if err != nil {
			t.Fatal(err)
		}

		if req.Method != "GET" {
			t.Errorf("Method = %q, want %q", req.Method, "GET")
		}
		if req.RequestURI != "/users" {
			t.Errorf("RequestURI = %q, want %q", req.RequestURI, "/users")
		}
		if req.Proto != "HTTP/1.1" {
			t.Errorf("Proto = %q, want %q", req.Proto, "HTTP/1.1")
		}
		if req.Host != "localhost" {
			t.Errorf("Host = %q, want %q", req.Host, "localhost")
		}
	})

	t.Run("POST with body", func(t *testing.T) {
		raw := "POST /submit HTTP/1.1\r\nHost: localhost\r\nContent-Length: 11\r\n\r\nHello World"
		br := bufio.NewReader(strings.NewReader(raw))
		req, err := ReadRequest(br)
		if err != nil {
			t.Fatal(err)
		}

		if req.Method != "POST" {
			t.Errorf("Method = %q, want %q", req.Method, "POST")
		}
		if req.ContentLength != 11 {
			t.Errorf("ContentLength = %d, want %d", req.ContentLength, 11)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != "Hello World" {
			t.Errorf("Body = %q, want %q", string(body), "Hello World")
		}
	})

	t.Run("multiple headers", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nHost: localhost\r\nAccept: text/html\r\nAccept-Encoding: gzip\r\n\r\n"
		br := bufio.NewReader(strings.NewReader(raw))
		req, err := ReadRequest(br)
		if err != nil {
			t.Fatal(err)
		}

		if req.Header.Get("Accept") != "text/html" {
			t.Errorf("Accept = %q, want %q", req.Header.Get("Accept"), "text/html")
		}
		if req.Header.Get("Accept-Encoding") != "gzip" {
			t.Errorf("Accept-Encoding = %q, want %q", req.Header.Get("Accept-Encoding"), "gzip")
		}
	})

	t.Run("no Content-Length", func(t *testing.T) {
		raw := "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"
		br := bufio.NewReader(strings.NewReader(raw))
		req, err := ReadRequest(br)
		if err != nil {
			t.Fatal(err)
		}
		if req.ContentLength != -1 {
			t.Errorf("ContentLength = %d, want %d", req.ContentLength, -1)
		}
	})

	t.Run("malformed request line", func(t *testing.T) {
		br := bufio.NewReader(strings.NewReader("INVALID\r\n\r\n"))
		_, err := ReadRequest(br)
		if err == nil {
			t.Error("expected error for malformed request line, got nil")
		}
	})

	t.Run("malformed protocol", func(t *testing.T) {
		br := bufio.NewReader(strings.NewReader("GET / BOGUS/1.1\r\nHost: localhost\r\n\r\n"))
		_, err := ReadRequest(br)
		if err == nil {
			t.Error("expected error for malformed protocol, got nil")
		}
	})

	t.Run("bad Content-Length", func(t *testing.T) {
		raw := "POST / HTTP/1.1\r\nHost: localhost\r\nContent-Length: abc\r\n\r\n"
		br := bufio.NewReader(strings.NewReader(raw))
		_, err := ReadRequest(br)
		if err == nil {
			t.Error("expected error for bad Content-Length, got nil")
		}
	})
}
