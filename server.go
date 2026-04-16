package bolt

import (
	"bufio"
	"net"
)

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
	Addr    string
	Handler Handler
}

// ListenAndServe listens on the TCP address s.Addr and serves incoming HTTP requests.
func (s *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	br := bufio.NewReader(conn)
	bw := bufio.NewWriter(conn)

	for {
		req, err := ReadRequest(br)
		if err != nil {
			return
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
