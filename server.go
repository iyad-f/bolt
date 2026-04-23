package bolt

import (
	"context"
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
	Addr         string
	Handler      Handler
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	listener     net.Listener
	activeConn   sync.WaitGroup
	shuttingDown atomic.Bool
}

// ListenAndServe listens on the TCP address s.Addr and serves incoming HTTP requests.
func (s *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	s.listener = listener

	for {
		conn, err := listener.Accept()
		if err != nil {
			if s.shuttingDown.Load() {
				return ErrServerClosed
			}
			return err
		}

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

	s.activeConn.Add(1)
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
