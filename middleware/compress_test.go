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
