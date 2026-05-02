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
