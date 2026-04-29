package middleware

import (
	"context"
	"strings"
	"testing"

	"github.com/iyad-f/bolt"
)

func TestRecovery(t *testing.T) {
	router := bolt.New()
	router.Use(Recovery())
	router.GET("/panic", func(w bolt.ResponseWriter, r *bolt.Request) {
		panic("test panic")
	})
	router.GET("/ok", func(w bolt.ResponseWriter, r *bolt.Request) {
		w.Write([]byte("ok"))
	})

	addr, server := startServer(t, router)
	defer server.Shutdown(context.Background())

	t.Run("recovers from panic", func(t *testing.T) {
		resp := doRequest(t, addr, "GET /panic HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "500") {
			t.Errorf("expected 500, got: %s", resp)
		}
		if !strings.Contains(resp, "Internal Server Error") {
			t.Errorf("expected 'Internal Server Error', got: %s", resp)
		}
	})

	t.Run("normal request unaffected", func(t *testing.T) {
		resp := doRequest(t, addr, "GET /ok HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "200 OK") {
			t.Errorf("expected 200, got: %s", resp)
		}
		if !strings.Contains(resp, "ok") {
			t.Errorf("expected 'ok' in body, got: %s", resp)
		}
	})
}
