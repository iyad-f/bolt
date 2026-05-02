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
	"bytes"
	"context"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/iyad-f/bolt"
)

func TestLogger(t *testing.T) {
	t.Run("logs request info", func(t *testing.T) {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr)

		router := bolt.New()
		router.Use(Logger())
		router.GET("/hello", func(w bolt.ResponseWriter, r *bolt.Request) {
			w.Write([]byte("hello"))
		})

		addr, server := startServer(t, router)
		defer server.Shutdown(context.Background())

		doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")

		logOutput := buf.String()
		if !strings.Contains(logOutput, "GET") {
			t.Errorf("expected 'GET' in log, got: %s", logOutput)
		}
		if !strings.Contains(logOutput, "/hello") {
			t.Errorf("expected '/hello' in log, got: %s", logOutput)
		}
		if !strings.Contains(logOutput, "200") {
			t.Errorf("expected '200' in log, got: %s", logOutput)
		}
	})
}
