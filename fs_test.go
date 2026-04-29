package bolt

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupFileServer(t *testing.T) (string, *Server, net.Listener) {
	t.Helper()

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello world"), 0644)
	os.WriteFile(filepath.Join(dir, "style.css"), []byte("body { color: red; }"), 0644)
	os.Mkdir(filepath.Join(dir, "sub"), 0755)
	os.WriteFile(filepath.Join(dir, "sub", "index.html"), []byte("<h1>Index</h1>"), 0644)

	router := New()
	router.Static("/static", dir)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	server := &Server{Handler: router}
	go server.Serve(listener)

	return listener.Addr().String(), server, listener
}

func TestFileServer(t *testing.T) {
	t.Run("serve file", func(t *testing.T) {
		addr, server, _ := setupFileServer(t)
		defer server.Shutdown(context.Background())

		conn, err := net.Dial("tcp", addr)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		fmt.Fprintf(conn, "GET /static/hello.txt HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
		resp, err := io.ReadAll(conn)
		if err != nil {
			t.Fatal(err)
		}

		body := string(resp)
		if !strings.Contains(body, "200 OK") {
			t.Errorf("expected 200, got: %s", body)
		}
		if !strings.Contains(body, "hello world") {
			t.Errorf("expected 'hello world' in body, got: %s", body)
		}
		if !strings.Contains(body, "Content-Type: text/plain") {
			t.Errorf("expected text/plain content type, got: %s", body)
		}
	})

	t.Run("serve css", func(t *testing.T) {
		addr, server, _ := setupFileServer(t)
		defer server.Shutdown(context.Background())

		conn, err := net.Dial("tcp", addr)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		fmt.Fprintf(conn, "GET /static/style.css HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
		resp, err := io.ReadAll(conn)
		if err != nil {
			t.Fatal(err)
		}

		body := string(resp)
		if !strings.Contains(body, "text/css") {
			t.Errorf("expected text/css content type, got: %s", body)
		}
	})

	t.Run("directory serves index.html", func(t *testing.T) {
		addr, server, _ := setupFileServer(t)
		defer server.Shutdown(context.Background())

		conn, err := net.Dial("tcp", addr)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		fmt.Fprintf(conn, "GET /static/sub HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
		resp, err := io.ReadAll(conn)
		if err != nil {
			t.Fatal(err)
		}

		body := string(resp)
		if !strings.Contains(body, "<h1>Index</h1>") {
			t.Errorf("expected index.html content, got: %s", body)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		addr, server, _ := setupFileServer(t)
		defer server.Shutdown(context.Background())

		conn, err := net.Dial("tcp", addr)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		fmt.Fprintf(conn, "GET /static/nonexistent.txt HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
		resp, err := io.ReadAll(conn)
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(string(resp), "404") {
			t.Errorf("expected 404, got: %s", string(resp))
		}
	})

	t.Run("path traversal blocked", func(t *testing.T) {
		addr, server, _ := setupFileServer(t)
		defer server.Shutdown(context.Background())

		conn, err := net.Dial("tcp", addr)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		fmt.Fprintf(conn, "GET /static/../../etc/passwd HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
		resp, err := io.ReadAll(conn)
		if err != nil {
			t.Fatal(err)
		}

		body := string(resp)
		if strings.Contains(body, "root:") {
			t.Error("path traversal was not blocked")
		}
	})

	t.Run("etag header present", func(t *testing.T) {
		addr, server, _ := setupFileServer(t)
		defer server.Shutdown(context.Background())

		conn, err := net.Dial("tcp", addr)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		fmt.Fprintf(conn, "GET /static/hello.txt HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
		resp, err := io.ReadAll(conn)
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(string(resp), "Etag:") {
			t.Errorf("expected Etag header, got: %s", string(resp))
		}
	})
}
