package middleware

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/iyad-f/bolt"
)

func TestRateLimit(t *testing.T) {
	router := bolt.New()
	router.Use(RateLimit(RateLimitConfig{
		Max:    3,
		Window: time.Minute,
	}))
	router.GET("/hello", func(w bolt.ResponseWriter, r *bolt.Request) {
		w.Write([]byte("hello"))
	})

	addr, server := startServer(t, router)
	defer server.Shutdown(context.Background())

	t.Run("allows requests under limit", func(t *testing.T) {
		resp := doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "200 OK") {
			t.Errorf("expected 200, got: %s", resp)
		}
		if !strings.Contains(resp, "hello") {
			t.Error("expected body 'hello'")
		}
	})

	t.Run("sets rate limit headers", func(t *testing.T) {
		resp := doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "X-Ratelimit-Limit: 3") {
			t.Errorf("expected X-Ratelimit-Limit: 3, got: %s", resp)
		}
		if !strings.Contains(resp, "X-Ratelimit-Remaining:") {
			t.Errorf("expected X-Ratelimit-Remaining header, got: %s", resp)
		}
		if !strings.Contains(resp, "X-Ratelimit-Reset:") {
			t.Errorf("expected X-Ratelimit-Reset header, got: %s", resp)
		}
	})

	t.Run("denies requests over limit", func(t *testing.T) {
		doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
		resp := doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "429") {
			t.Errorf("expected 429, got: %s", resp)
		}
		if !strings.Contains(resp, "Too Many Requests") {
			t.Errorf("expected 'Too Many Requests', got: %s", resp)
		}
		if !strings.Contains(resp, "Retry-After:") {
			t.Errorf("expected Retry-After header, got: %s", resp)
		}
	})
}

func TestRateLimitPerClient(t *testing.T) {
	store := newMemoryStore(2, time.Minute)

	router := bolt.New()
	router.Use(RateLimit(RateLimitConfig{
		Max:    2,
		Window: time.Minute,
		Store:  store,
		KeyFunc: func(r *bolt.Request) string {
			return r.Header.Get("X-Client-ID")
		},
	}))
	router.GET("/hello", func(w bolt.ResponseWriter, r *bolt.Request) {
		w.Write([]byte("hello"))
	})

	addr, server := startServer(t, router)
	defer server.Shutdown(context.Background())

	t.Run("different clients have separate limits", func(t *testing.T) {
		// client A: 2 requests
		resp := doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nX-Client-ID: client-a\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "200 OK") {
			t.Errorf("client-a request 1 should be allowed, got: %s", resp)
		}
		resp = doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nX-Client-ID: client-a\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "200 OK") {
			t.Errorf("client-a request 2 should be allowed, got: %s", resp)
		}

		// client A: 3rd request should be denied
		resp = doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nX-Client-ID: client-a\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "429") {
			t.Errorf("client-a request 3 should be denied, got: %s", resp)
		}

		// client B: should still be allowed
		resp = doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nX-Client-ID: client-b\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "200 OK") {
			t.Errorf("client-b should be allowed, got: %s", resp)
		}
	})
}

func TestRateLimitCustomDenyHandler(t *testing.T) {
	router := bolt.New()
	router.Use(RateLimit(RateLimitConfig{
		Max:    1,
		Window: time.Minute,
		DenyHandler: func(w bolt.ResponseWriter, r *bolt.Request) {
			w.WriteHeader(bolt.StatusTooManyRequests)
			w.Write([]byte("custom denied"))
		},
	}))
	router.GET("/hello", func(w bolt.ResponseWriter, r *bolt.Request) {
		w.Write([]byte("hello"))
	})

	addr, server := startServer(t, router)
	defer server.Shutdown(context.Background())

	// first request allowed
	doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")

	t.Run("uses custom deny handler", func(t *testing.T) {
		resp := doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "custom denied") {
			t.Errorf("expected custom deny body, got: %s", resp)
		}
	})
}

type errorStore struct{}

func (e *errorStore) Allow(key string) (int, time.Time, bool, error) {
	return 0, time.Time{}, false, errors.New("store error")
}

func TestRateLimitErrorHandler(t *testing.T) {
	router := bolt.New()
	router.Use(RateLimit(RateLimitConfig{
		Max:    10,
		Window: time.Minute,
		Store:  &errorStore{},
	}))
	router.GET("/hello", func(w bolt.ResponseWriter, r *bolt.Request) {
		w.Write([]byte("hello"))
	})

	addr, server := startServer(t, router)
	defer server.Shutdown(context.Background())

	t.Run("returns 500 on store error", func(t *testing.T) {
		resp := doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "500") {
			t.Errorf("expected 500, got: %s", resp)
		}
		if !strings.Contains(resp, "Internal Server Error") {
			t.Errorf("expected 'Internal Server Error', got: %s", resp)
		}
	})
}

func TestRateLimitCustomErrorHandler(t *testing.T) {
	router := bolt.New()
	router.Use(RateLimit(RateLimitConfig{
		Max:    10,
		Window: time.Minute,
		Store:  &errorStore{},
		ErrorHandler: func(w bolt.ResponseWriter, r *bolt.Request, err error) {
			w.WriteHeader(bolt.StatusServiceUnavailable)
			w.Write([]byte("custom error"))
		},
	}))
	router.GET("/hello", func(w bolt.ResponseWriter, r *bolt.Request) {
		w.Write([]byte("hello"))
	})

	addr, server := startServer(t, router)
	defer server.Shutdown(context.Background())

	t.Run("uses custom error handler", func(t *testing.T) {
		resp := doRequest(t, addr, "GET /hello HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
		if !strings.Contains(resp, "503") {
			t.Errorf("expected 503, got: %s", resp)
		}
		if !strings.Contains(resp, "custom error") {
			t.Errorf("expected custom error body, got: %s", resp)
		}
	})
}

func TestMemoryStoreSlidingWindow(t *testing.T) {
	t.Run("allows up to max requests", func(t *testing.T) {
		store := newMemoryStore(3, time.Minute)
		for i := 0; i < 3; i++ {
			remaining, _, allowed, err := store.Allow("client1")
			if err != nil {
				t.Fatal(err)
			}
			if !allowed {
				t.Errorf("request %d should be allowed", i+1)
			}
			if remaining != 3-i-1 {
				t.Errorf("request %d: expected remaining %d, got %d", i+1, 3-i-1, remaining)
			}
		}

		// 4th request should be denied
		remaining, _, allowed, err := store.Allow("client1")
		if err != nil {
			t.Fatal(err)
		}
		if allowed {
			t.Error("4th request should be denied")
		}
		if remaining != 0 {
			t.Errorf("expected remaining 0, got %d", remaining)
		}
	})

	t.Run("separate keys are independent", func(t *testing.T) {
		store := newMemoryStore(1, time.Minute)

		_, _, allowed, _ := store.Allow("a")
		if !allowed {
			t.Error("key 'a' first request should be allowed")
		}

		_, _, allowed, _ = store.Allow("a")
		if allowed {
			t.Error("key 'a' second request should be denied")
		}

		_, _, allowed, _ = store.Allow("b")
		if !allowed {
			t.Error("key 'b' first request should be allowed")
		}
	})

	t.Run("resets after window expires", func(t *testing.T) {
		store := newMemoryStore(1, 50*time.Millisecond)

		_, _, allowed, _ := store.Allow("client")
		if !allowed {
			t.Error("first request should be allowed")
		}

		_, _, allowed, _ = store.Allow("client")
		if allowed {
			t.Error("second request should be denied")
		}

		// wait for window to expire
		time.Sleep(110 * time.Millisecond)

		_, _, allowed, _ = store.Allow("client")
		if !allowed {
			t.Error("request after window reset should be allowed")
		}
	})
}
