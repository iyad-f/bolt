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
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	t.Run("full request response", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}

		server := &Server{Handler: HandlerFunc(func(w ResponseWriter, r *Request) {
			w.Write([]byte("hello"))
		})}
		go server.Serve(listener)
		defer server.Shutdown(context.Background())

		conn, err := net.Dial("tcp", listener.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		fmt.Fprintf(conn, "GET / HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")

		resp, err := io.ReadAll(conn)
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(string(resp), "HTTP/1.1 200 OK") {
			t.Errorf("expected 200 OK, got: %s", string(resp))
		}
		if !strings.Contains(string(resp), "hello") {
			t.Errorf("expected 'hello' in body, got: %s", string(resp))
		}
	})

	t.Run("keep alive", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}

		count := 0
		server := &Server{Handler: HandlerFunc(func(w ResponseWriter, r *Request) {
			count++
			fmt.Fprintf(w, "request %d", count)
		})}
		go server.Serve(listener)
		defer server.Shutdown(context.Background())

		conn, err := net.Dial("tcp", listener.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)

		fmt.Fprintf(conn, "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n")
		resp1, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(resp1, "200 OK") {
			t.Errorf("request 1: expected 200 OK, got: %s", resp1)
		}

		for {
			line, _ := reader.ReadString('\n')
			if line == "\r\n" {
				break
			}
		}
		body1 := make([]byte, 9)
		io.ReadFull(reader, body1)
		if string(body1) != "request 1" {
			t.Errorf("request 1: body = %q, want %q", string(body1), "request 1")
		}

		fmt.Fprintf(conn, "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n")
		resp2, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(resp2, "200 OK") {
			t.Errorf("request 2: expected 200 OK, got: %s", resp2)
		}
	})

	t.Run("graceful shutdown", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}

		handlerStarted := make(chan struct{})
		server := &Server{Handler: HandlerFunc(func(w ResponseWriter, r *Request) {
			close(handlerStarted)
			time.Sleep(100 * time.Millisecond)
			w.Write([]byte("done"))
		})}
		go server.Serve(listener)

		conn, err := net.Dial("tcp", listener.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		fmt.Fprintf(conn, "GET / HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")

		<-handlerStarted

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		errCh := make(chan error, 1)
		go func() {
			errCh <- server.Shutdown(ctx)
		}()

		resp, err := io.ReadAll(conn)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(resp), "done") {
			t.Errorf("expected 'done' in body, got: %s", string(resp))
		}

		if err := <-errCh; err != nil {
			t.Errorf("shutdown error: %v", err)
		}
	})
}
