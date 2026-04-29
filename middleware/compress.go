package middleware

import (
	"compress/gzip"
	"strings"

	"github.com/iyad-f/bolt"
)

type gzipWriter struct {
	bolt.ResponseWriter
	gz *gzip.Writer
}

func (g *gzipWriter) Write(p []byte) (int, error) {
	return g.gz.Write(p)
}

// Compress returns a middleware that gzip-compresses response bodies when the client supports it.
func Compress() bolt.MiddlewareFunc {
	return func(next bolt.Handler) bolt.Handler {
		return bolt.HandlerFunc(func(w bolt.ResponseWriter, r *bolt.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			gz := gzip.NewWriter(w)
			defer gz.Close()
			gw := &gzipWriter{ResponseWriter: w, gz: gz}
			w.Header().Set("Content-Encoding", "gzip")
			next.ServeHTTP(gw, r)
		})
	}
}
