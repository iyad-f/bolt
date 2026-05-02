package middleware

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/iyad-f/bolt"
)

// RateLimitStore is the interface for rate limit backends. Implement this
// to use a custom storage such as Redis.
type RateLimitStore interface {
	Allow(key string) (remaining int, resetTime time.Time, allowed bool, err error)
}

// RateLimitConfig defines the configuration for the RateLimit middleware.
type RateLimitConfig struct {
	// Max is the maximum number of requests allowed per window. Defaults to 100.
	Max int

	// Window is the duration of the rate limit window. Defaults to 1 minute.
	Window time.Duration

	// KeyFunc extracts a key from the request to identify the client. Defaults to client IP.
	KeyFunc func(r *bolt.Request) string

	// Store is the backend used to track request counts. Defaults to an in-memory
	// sliding window counter.
	Store RateLimitStore

	// DenyHandler is called when a request is rate limited. Defaults to 429 Too Many Requests.
	DenyHandler func(w bolt.ResponseWriter, r *bolt.Request)

	// ErrorHandler is called when the store returns an error. Defaults to 500 Internal Server Error.
	ErrorHandler func(w bolt.ResponseWriter, r *bolt.Request, err error)
}

func keyByIP(r *bolt.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// DefaultRateLimitConfig returns a RateLimitConfig with 100 requests per minute, keyed by client IP.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Max:     100,
		Window:  time.Minute,
		KeyFunc: keyByIP,
	}
}

type slidingWindowEntry struct {
	currCount int
	prevCount int
	currStart time.Time
}

type memoryStore struct {
	mu      sync.Mutex
	max     int
	window  time.Duration
	entries map[string]*slidingWindowEntry
}

func newMemoryStore(max int, window time.Duration) *memoryStore {
	return &memoryStore{
		max:     max,
		window:  window,
		entries: make(map[string]*slidingWindowEntry),
	}
}

func (m *memoryStore) Allow(key string) (remaining int, resetTime time.Time, allowed bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.entries[key]
	now := time.Now()
	if !ok {
		entry = &slidingWindowEntry{
			currCount: 0,
			prevCount: 0,
			currStart: now,
		}
		m.entries[key] = entry
	}

	// client was idle for 2+ windows, so behave as if its a new client
	if now.Sub(entry.currStart) >= 2*m.window {
		entry.prevCount = 0
		entry.currCount = 0
		entry.currStart = now
	} else if now.Sub(entry.currStart) >= m.window {
		entry.prevCount = entry.currCount
		entry.currCount = 0
		entry.currStart = entry.currStart.Add(m.window)
	}

	timeLeft := entry.currStart.Add(m.window).Sub(now)
	weight := float64(timeLeft) / float64(m.window)
	effectiveCount := float64(entry.prevCount)*weight + float64(entry.currCount)

	if effectiveCount >= float64(m.max) {
		return 0, entry.currStart.Add(m.window), false, nil
	}

	resetTime = entry.currStart.Add(m.window)
	entry.currCount++
	remaining = m.max - int(effectiveCount) - 1

	return remaining, resetTime, true, nil
}

// RateLimit returns a middleware that limits requests using a sliding window counter.
func RateLimit(config RateLimitConfig) bolt.MiddlewareFunc {
	if config.Max == 0 {
		config.Max = 100
	}

	if config.Window == 0 {
		config.Window = time.Minute
	}

	if config.KeyFunc == nil {
		config.KeyFunc = keyByIP
	}

	if config.Store == nil {
		config.Store = newMemoryStore(config.Max, config.Window)
	}

	if config.DenyHandler == nil {
		config.DenyHandler = func(w bolt.ResponseWriter, r *bolt.Request) {
			w.WriteHeader(bolt.StatusTooManyRequests)
			w.Write([]byte(bolt.StatusText(bolt.StatusTooManyRequests)))
		}
	}

	if config.ErrorHandler == nil {
		config.ErrorHandler = func(w bolt.ResponseWriter, r *bolt.Request, err error) {
			w.WriteHeader(bolt.StatusInternalServerError)
			w.Write([]byte(bolt.StatusText(bolt.StatusInternalServerError)))
		}
	}

	return func(next bolt.Handler) bolt.Handler {
		return bolt.HandlerFunc(func(w bolt.ResponseWriter, r *bolt.Request) {
			key := config.KeyFunc(r)
			remaining, resetTime, allowed, err := config.Store.Allow(key)
			if err != nil {
				config.ErrorHandler(w, r, err)
				return
			}

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.Max))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(resetTime.Unix())))

			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(int(time.Until(resetTime).Seconds())))
				config.DenyHandler(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
