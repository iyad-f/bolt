package middleware

import (
	"context"
	"strings"
	"testing"

	"github.com/iyad-f/bolt"
)

func TestCORS(t *testing.T) {
	router := bolt.New()
	router.Use(CORS(DefaultCORSConfig()))
	router.GET("/hello", func(w bolt.ResponseWriter, r *bolt.Request) {
		w.Write([]byte("hello"))
	})

	addr, server := startServer(t, router)
	defer server.Shutdown(context.Background())

	t.Run("no origin header skips CORS", func(t *testing.T) {
		resp := doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
		if strings.Contains(resp, "Access-Control-Allow-Origin") {
			t.Error("should not set CORS headers without Origin")
		}
		if !strings.Contains(resp, "hello") {
			t.Error("expected body 'hello'")
		}
	})

	t.Run("simple request with origin", func(t *testing.T) {
		resp := doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nOrigin: http://example.com\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "Access-Control-Allow-Origin: *") {
			t.Errorf("expected Access-Control-Allow-Origin: *, got: %s", resp)
		}
		if !strings.Contains(resp, "Vary: Origin") {
			t.Errorf("expected Vary: Origin, got: %s", resp)
		}
		if !strings.Contains(resp, "hello") {
			t.Error("expected body 'hello'")
		}
	})

	t.Run("preflight request", func(t *testing.T) {
		resp := doRequest(t, addr, "OPTIONS /hello HTTP/1.1\r\nHost: localhost\r\nOrigin: http://example.com\r\nAccess-Control-Request-Method: POST\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "204") {
			t.Errorf("expected 204, got: %s", resp)
		}
		if !strings.Contains(resp, "Access-Control-Allow-Origin: *") {
			t.Errorf("expected Access-Control-Allow-Origin: *, got: %s", resp)
		}
		if !strings.Contains(resp, "Access-Control-Allow-Methods: GET, POST, HEAD") {
			t.Errorf("expected Access-Control-Allow-Methods, got: %s", resp)
		}
	})

	t.Run("OPTIONS without request method is not preflight", func(t *testing.T) {
		resp := doRequest(t, addr, "OPTIONS /hello HTTP/1.1\r\nHost: localhost\r\nOrigin: http://example.com\r\nConnection: close\r\n\r\n")
		if strings.Contains(resp, "204") {
			t.Error("bare OPTIONS should not be treated as preflight")
		}
		if !strings.Contains(resp, "Access-Control-Allow-Origin: *") {
			t.Errorf("should still set Allow-Origin, got: %s", resp)
		}
	})
}

func TestCORSSpecificOrigins(t *testing.T) {
	router := bolt.New()
	router.Use(CORS(CORSConfig{
		AllowedOrigins: []string{"http://allowed.com", "http://also-allowed.com"},
		AllowedMethods: []string{"GET", "POST"},
	}))
	router.GET("/hello", func(w bolt.ResponseWriter, r *bolt.Request) {
		w.Write([]byte("hello"))
	})

	addr, server := startServer(t, router)
	defer server.Shutdown(context.Background())

	t.Run("allowed origin is echoed back", func(t *testing.T) {
		resp := doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nOrigin: http://allowed.com\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "Access-Control-Allow-Origin: http://allowed.com") {
			t.Errorf("expected origin echo, got: %s", resp)
		}
	})

	t.Run("disallowed origin gets no CORS headers", func(t *testing.T) {
		resp := doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nOrigin: http://evil.com\r\nConnection: close\r\n\r\n")
		if strings.Contains(resp, "Access-Control-Allow-Origin") {
			t.Error("should not set Allow-Origin for disallowed origin")
		}
		if !strings.Contains(resp, "hello") {
			t.Error("should still serve the response body")
		}
	})
}

func TestCORSWildcardOrigin(t *testing.T) {
	router := bolt.New()
	router.Use(CORS(CORSConfig{
		AllowedOrigins: []string{"http://*.example.com"},
		AllowedMethods: []string{"GET"},
	}))
	router.GET("/hello", func(w bolt.ResponseWriter, r *bolt.Request) {
		w.Write([]byte("hello"))
	})

	addr, server := startServer(t, router)
	defer server.Shutdown(context.Background())

	t.Run("matches wildcard origin", func(t *testing.T) {
		resp := doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nOrigin: http://sub.example.com\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "Access-Control-Allow-Origin: http://sub.example.com") {
			t.Errorf("expected wildcard match, got: %s", resp)
		}
	})

	t.Run("rejects non-matching origin", func(t *testing.T) {
		resp := doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nOrigin: http://other.com\r\nConnection: close\r\n\r\n")
		if strings.Contains(resp, "Access-Control-Allow-Origin") {
			t.Error("should not match non-matching origin")
		}
	})

	t.Run("rejects empty subdomain", func(t *testing.T) {
		resp := doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nOrigin: http://.example.com\r\nConnection: close\r\n\r\n")
		if strings.Contains(resp, "Access-Control-Allow-Origin") {
			t.Error("should not match empty subdomain")
		}
	})
}

func TestCORSCredentials(t *testing.T) {
	router := bolt.New()
	router.Use(CORS(CORSConfig{
		AllowedOrigins:   []string{"http://example.com"},
		AllowCredentials: true,
	}))
	router.GET("/hello", func(w bolt.ResponseWriter, r *bolt.Request) {
		w.Write([]byte("hello"))
	})

	addr, server := startServer(t, router)
	defer server.Shutdown(context.Background())

	t.Run("sets credentials header", func(t *testing.T) {
		resp := doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nOrigin: http://example.com\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "Access-Control-Allow-Credentials: true") {
			t.Errorf("expected Allow-Credentials, got: %s", resp)
		}
	})
}

func TestCORSWildcardWithCredentialsPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for wildcard + credentials")
		}
	}()

	CORS(CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})
}

func TestCORSExposeHeaders(t *testing.T) {
	router := bolt.New()
	router.Use(CORS(CORSConfig{
		AllowedOrigins: []string{"*"},
		ExposedHeaders: []string{"X-Custom-Header", "X-Another"},
	}))
	router.GET("/hello", func(w bolt.ResponseWriter, r *bolt.Request) {
		w.Write([]byte("hello"))
	})

	addr, server := startServer(t, router)
	defer server.Shutdown(context.Background())

	t.Run("sets expose headers on actual request", func(t *testing.T) {
		resp := doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nOrigin: http://example.com\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "Access-Control-Expose-Headers: X-Custom-Header, X-Another") {
			t.Errorf("expected Expose-Headers, got: %s", resp)
		}
	})
}

func TestCORSMaxAge(t *testing.T) {
	router := bolt.New()
	router.Use(CORS(CORSConfig{
		AllowedOrigins: []string{"*"},
		MaxAge:         3600,
	}))
	router.GET("/hello", func(w bolt.ResponseWriter, r *bolt.Request) {
		w.Write([]byte("hello"))
	})

	addr, server := startServer(t, router)
	defer server.Shutdown(context.Background())

	t.Run("sets max age on preflight", func(t *testing.T) {
		resp := doRequest(t, addr, "OPTIONS /hello HTTP/1.1\r\nHost: localhost\r\nOrigin: http://example.com\r\nAccess-Control-Request-Method: POST\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "Access-Control-Max-Age: 3600") {
			t.Errorf("expected Max-Age: 3600, got: %s", resp)
		}
	})
}

func TestCORSAllowOriginFunc(t *testing.T) {
	router := bolt.New()
	router.Use(CORS(CORSConfig{
		AllowedOrigins: []string{"http://static.com"},
		AllowOriginFunc: func(origin string) bool {
			return strings.HasSuffix(origin, ".dynamic.com")
		},
	}))
	router.GET("/hello", func(w bolt.ResponseWriter, r *bolt.Request) {
		w.Write([]byte("hello"))
	})

	addr, server := startServer(t, router)
	defer server.Shutdown(context.Background())

	t.Run("static origin still works", func(t *testing.T) {
		resp := doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nOrigin: http://static.com\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "Access-Control-Allow-Origin: http://static.com") {
			t.Errorf("expected static origin match, got: %s", resp)
		}
	})

	t.Run("func allows dynamic origin", func(t *testing.T) {
		resp := doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nOrigin: http://app.dynamic.com\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "Access-Control-Allow-Origin: http://app.dynamic.com") {
			t.Errorf("expected func match, got: %s", resp)
		}
	})

	t.Run("func rejects non-matching origin", func(t *testing.T) {
		resp := doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nOrigin: http://evil.com\r\nConnection: close\r\n\r\n")
		if strings.Contains(resp, "Access-Control-Allow-Origin") {
			t.Error("should not allow non-matching origin")
		}
	})
}

func TestCORSPreflightAllowHeaders(t *testing.T) {
	router := bolt.New()
	router.Use(CORS(CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"X-Custom", "Authorization"},
	}))
	router.GET("/hello", func(w bolt.ResponseWriter, r *bolt.Request) {
		w.Write([]byte("hello"))
	})

	addr, server := startServer(t, router)
	defer server.Shutdown(context.Background())

	t.Run("preflight includes allowed headers", func(t *testing.T) {
		resp := doRequest(t, addr, "OPTIONS /hello HTTP/1.1\r\nHost: localhost\r\nOrigin: http://example.com\r\nAccess-Control-Request-Method: POST\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "Access-Control-Allow-Headers: x-custom, authorization") {
			t.Errorf("expected Allow-Headers (lowercased), got: %s", resp)
		}
	})
}
