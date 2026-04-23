package middleware

import (
	"log"
	"time"

	"github.com/iyad-f/bolt"
)

type responseRecorder struct {
	bolt.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(statusCode int) {
	rr.statusCode = statusCode
	rr.ResponseWriter.WriteHeader(statusCode)
}

// Logger returns a middleware that logs each request's method, path, status code, and duration.
func Logger() bolt.MiddlewareFunc {
	return func(next bolt.Handler) bolt.Handler {
		return bolt.HandlerFunc(func(w bolt.ResponseWriter, r *bolt.Request) {
			start := time.Now()
			rr := &responseRecorder{ResponseWriter: w, statusCode: bolt.StatusOK}
			next.ServeHTTP(rr, r)
			log.Printf("%s %s %d %s", r.Method, r.URL.Path, rr.statusCode, time.Since(start))
		})
	}
}
