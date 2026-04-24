package bolt

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestResponse(t *testing.T) {
	t.Run("write defaults to 200", func(t *testing.T) {
		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)
		resp := &response{writer: bw, header: Header{}}

		resp.Write([]byte("hello"))
		resp.flush()

		output := buf.String()
		if !strings.Contains(output, "HTTP/1.1 200 OK") {
			t.Errorf("expected status 200, got: %s", output)
		}
	})

	t.Run("write header sets status", func(t *testing.T) {
		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)
		resp := &response{writer: bw, header: Header{}}

		resp.WriteHeader(StatusNotFound)
		resp.flush()

		output := buf.String()
		if !strings.Contains(output, "HTTP/1.1 404 Not Found") {
			t.Errorf("expected status 404, got: %s", output)
		}
	})

	t.Run("write header only writes once", func(t *testing.T) {
		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)
		resp := &response{writer: bw, header: Header{}}

		resp.WriteHeader(StatusNotFound)
		resp.WriteHeader(StatusInternalServerError)
		resp.flush()

		output := buf.String()
		if !strings.Contains(output, "HTTP/1.1 404 Not Found") {
			t.Errorf("expected status 404, got: %s", output)
		}
	})

	t.Run("body buffering", func(t *testing.T) {
		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)
		resp := &response{writer: bw, header: Header{}}

		resp.Write([]byte("hello"))
		resp.Write([]byte(" world"))
		resp.flush()

		output := buf.String()
		if !strings.Contains(output, "hello world") {
			t.Errorf("response does not contain 'hello world'")
		}
	})

	t.Run("auto content-length", func(t *testing.T) {
		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)
		resp := &response{writer: bw, header: Header{}}

		resp.Write([]byte("hello"))
		resp.flush()

		output := buf.String()
		if !strings.Contains(output, "Content-Length: 5") {
			t.Errorf("expected Content-Length: 5, got: %s", output)
		}
	})
}
