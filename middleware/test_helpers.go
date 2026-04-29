package middleware

import (
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/iyad-f/bolt"
)

func startServer(t *testing.T, handler bolt.Handler) (string, *bolt.Server) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	server := &bolt.Server{Handler: handler}
	go server.Serve(listener)

	return listener.Addr().String(), server
}

func doRequest(t *testing.T, addr, rawReq string) string {
	t.Helper()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	fmt.Fprint(conn, rawReq)
	resp, err := io.ReadAll(conn)
	if err != nil {
		t.Fatal(err)
	}
	return string(resp)
}
