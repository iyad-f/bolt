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
