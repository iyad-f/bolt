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
	"io"
	"strings"
	"testing"
)

func TestChunkReader(t *testing.T) {
	t.Run("single chunk", func(t *testing.T) {
		raw := "5\r\nHello\r\n0\r\n\r\n"

		br := bufio.NewReader(strings.NewReader(raw))
		cr := &chunkReader{r: br}

		body, err := io.ReadAll(cr)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != "Hello" {
			t.Errorf("body = %q, want %q", string(body), "Hello")
		}
	})

	t.Run("multiple chunks", func(t *testing.T) {
		raw := "5\r\nHello\r\n6\r\n World\r\n0\r\n\r\n"
		br := bufio.NewReader(strings.NewReader(raw))
		cr := &chunkReader{r: br}

		body, err := io.ReadAll(cr)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != "Hello World" {
			t.Errorf("body = %q, want %q", string(body), "Hello World")
		}
	})

	t.Run("empty body", func(t *testing.T) {
		raw := "0\r\n\r\n"
		br := bufio.NewReader(strings.NewReader(raw))
		cr := &chunkReader{r: br}

		body, err := io.ReadAll(cr)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != "" {
			t.Errorf("body = %q, want empty", string(body))
		}
	})
}

func TestChunkWriter(t *testing.T) {
	t.Run("single write", func(t *testing.T) {
		var buf bytes.Buffer
		cw := &chunkWriter{w: &buf}

		cw.Write([]byte("Hello"))

		want := "5\r\nHello\r\n"
		if buf.String() != want {
			t.Errorf("output = %q, want %q", buf.String(), want)
		}
	})

	t.Run("close", func(t *testing.T) {
		var buf bytes.Buffer
		cw := &chunkWriter{w: &buf}

		cw.Write([]byte("Hello"))
		cw.Close()

		want := "5\r\nHello\r\n0\r\n\r\n"
		if buf.String() != want {
			t.Errorf("output = %q, want %q", buf.String(), want)
		}
	})
}
