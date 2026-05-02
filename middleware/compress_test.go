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

package middleware

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"

	"github.com/iyad-f/bolt"
)

func TestCompress(t *testing.T) {
	router := bolt.New()
	router.Use(Compress())
	router.GET("/hello", func(w bolt.ResponseWriter, r *bolt.Request) {
		w.Write([]byte("hello world"))
	})

	addr, server := startServer(t, router)
	defer server.Shutdown(context.Background())

	t.Run("compresses when client accepts gzip", func(t *testing.T) {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		fmt.Fprint(conn, "GET /hello HTTP/1.1\r\nHost: localhost\r\nAccept-Encoding: gzip\r\nConnection: close\r\n\r\n")

		resp, err := io.ReadAll(conn)
		if err != nil {
			t.Fatal(err)
		}

		rawResp := string(resp)
		if !strings.Contains(rawResp, "Content-Encoding: gzip") {
			t.Errorf("expected Content-Encoding: gzip, got: %s", rawResp)
		}

		bodyIdx := strings.Index(rawResp, "\r\n\r\n")
		if bodyIdx == -1 {
			t.Fatal("no body separator found")
		}
		body := resp[bodyIdx+4:]

		gz, err := gzip.NewReader(strings.NewReader(string(body)))
		if err != nil {
			t.Fatalf("failed to create gzip reader: %v", err)
		}
		defer gz.Close()

		decompressed, err := io.ReadAll(gz)
		if err != nil {
			t.Fatalf("failed to decompress: %v", err)
		}

		if string(decompressed) != "hello world" {
			t.Errorf("decompressed body = %q, want %q", string(decompressed), "hello world")
		}
	})

	t.Run("no compression without accept-encoding", func(t *testing.T) {
		resp := doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
		if strings.Contains(resp, "Content-Encoding: gzip") {
			t.Error("should not compress without Accept-Encoding: gzip")
		}
		if !strings.Contains(resp, "hello world") {
			t.Errorf("expected 'hello world' in body, got: %s", resp)
		}
	})
}
