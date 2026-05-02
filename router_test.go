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
	"net/url"
	"strings"
	"testing"
)

func TestRouter(t *testing.T) {
	t.Run("route matching", func(t *testing.T) {
		router := New()
		router.GET("/hello", func(w ResponseWriter, r *Request) {
			w.Write([]byte("hello"))
		})

		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)
		resp := &response{writer: bw, header: Header{}}
		req := &Request{Method: "GET", URL: &url.URL{Path: "/hello"}}

		router.ServeHTTP(resp, req)
		resp.flush()

		if !strings.Contains(buf.String(), "hello") {
			t.Errorf("body does not contain 'hello'")
		}
	})

	t.Run("not found", func(t *testing.T) {
		router := New()
		router.GET("/hello", func(w ResponseWriter, r *Request) {
			w.Write([]byte("hello"))
		})

		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)
		resp := &response{writer: bw, header: Header{}}
		req := &Request{Method: "GET", URL: &url.URL{Path: "/nonexistent"}}

		router.ServeHTTP(resp, req)
		resp.flush()

		if !strings.Contains(buf.String(), "404 Not Found") {
			t.Errorf("expected 404, got: %s", buf.String())
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		router := New()
		router.GET("/hello", func(w ResponseWriter, r *Request) {
			w.Write([]byte("hello"))
		})

		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)
		resp := &response{writer: bw, header: Header{}}
		req := &Request{Method: "POST", URL: &url.URL{Path: "/hello"}}

		router.ServeHTTP(resp, req)
		resp.flush()

		if !strings.Contains(buf.String(), "405 Method Not Allowed") {
			t.Errorf("expected 405, got: %s", buf.String())
		}
	})

	t.Run("param extraction", func(t *testing.T) {
		router := New()
		router.GET("/users/:id", func(w ResponseWriter, r *Request) {
			w.Write([]byte("id=" + r.PathValue("id")))
		})

		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)
		resp := &response{writer: bw, header: Header{}}
		req := &Request{Method: "GET", URL: &url.URL{Path: "/users/42"}}

		router.ServeHTTP(resp, req)
		resp.flush()

		if !strings.Contains(buf.String(), "id=42") {
			t.Errorf("expected 'id=42' in body, got: %s", buf.String())
		}
	})

	t.Run("middleware", func(t *testing.T) {
		router := New()
		router.Use(func(next Handler) Handler {
			return HandlerFunc(func(w ResponseWriter, r *Request) {
				w.Header().Set("X-Test", "applied")
				next.ServeHTTP(w, r)
			})
		})
		router.GET("/hello", func(w ResponseWriter, r *Request) {
			w.Write([]byte("hello"))
		})

		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)
		resp := &response{writer: bw, header: Header{}}
		req := &Request{Method: "GET", URL: &url.URL{Path: "/hello"}}

		router.ServeHTTP(resp, req)
		resp.flush()

		output := buf.String()
		if !strings.Contains(output, "X-Test") {
			t.Errorf("middleware header not found in output: %s", output)
		}
		if !strings.Contains(output, "hello") {
			t.Errorf("body does not contain 'hello'")
		}
	})
}
