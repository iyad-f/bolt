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
	"context"
	"crypto/tls"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// ErrServerClosed is returned by ListenAndServe after a call to Shutdown.
var ErrServerClosed = errors.New("bolt: server closed")

// Handler responds to an HTTP request.
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}

// HandlerFunc is an adapter to allow the use of ordinary functions as HTTP handlers.
type HandlerFunc func(ResponseWriter, *Request)

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
	f(w, r)
}

// Server defines parameters for running an HTTP server.
type Server struct {
	// Addr is the TCP address to listen on, in the form "host:port".
	Addr string
	// Handler is the handler to invoke for incoming requests.
	Handler Handler
	// ReadTimeout is the maximum duration for reading the entire request.
	ReadTimeout time.Duration
	// WriteTimeout is the maximum duration before timing out writes of the response.
	WriteTimeout time.Duration
	// IdleTimeout is the maximum duration to wait for the next request on a keep-alive connection.
	IdleTimeout  time.Duration
	listener     net.Listener
	activeConn   sync.WaitGroup
	shuttingDown atomic.Bool
}

type onceCloseListener struct {
	net.Listener
	once sync.Once
	err  error
}

func (l *onceCloseListener) Close() error {
	l.once.Do(func() {
		l.err = l.Listener.Close()
	})
	return l.err
}

// ListenAndServe listens on the TCP address s.Addr and serves incoming HTTP requests.
func (s *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	return s.Serve(listener)
}

// ListenAndServeTLS listens on the TCP address s.Addr and serves HTTPS requests
// using the provided certificate and key files.
func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	listener = tls.NewListener(listener, &tls.Config{Certificates: []tls.Certificate{certificate}})
	return s.Serve(listener)
}

// Serve accepts incoming HTTP connections on the given listener.
func (s *Server) Serve(listener net.Listener) error {
	s.listener = &onceCloseListener{Listener: listener}
	defer s.listener.Close()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.shuttingDown.Load() {
				return ErrServerClosed
			}
			return err
		}

		s.activeConn.Add(1)
		go s.handleConn(conn)
	}
}

// Shutdown gracefully stops the server. It closes the listener and waits
// for active connections to finish, or until the context expires.
func (s *Server) Shutdown(ctx context.Context) error {
	s.shuttingDown.Store(true)
	closeErr := s.listener.Close()

	done := make(chan struct{})
	go func() {
		s.activeConn.Wait()
		close(done)
	}()

	select {
	case <-done:
		return closeErr
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	defer s.activeConn.Done()

	br := getReader(conn)
	bw := getWriter(conn)
	defer putReader(br)
	defer putWriter(bw)

	firstRequest := true

	for {
		if !firstRequest && s.IdleTimeout > 0 {
			conn.SetReadDeadline(time.Now().Add(s.IdleTimeout))
		}
		firstRequest = false

		req, err := ReadRequest(br)
		if err != nil {
			return
		}

		if s.ReadTimeout > 0 {
			conn.SetReadDeadline(time.Now().Add(s.ReadTimeout))
		}

		if s.WriteTimeout > 0 {
			conn.SetWriteDeadline(time.Now().Add(s.WriteTimeout))
		}

		resp := &response{writer: bw, header: Header{}}
		s.Handler.ServeHTTP(resp, req)

		if err := resp.flush(); err != nil {
			return
		}

		if req.Header.Get("Connection") == "close" {
			return
		}
	}
}
