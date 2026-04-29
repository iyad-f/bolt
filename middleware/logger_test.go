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
