package middleware

import "github.com/iyad-f/bolt"

// Recovery returns a middleware that recovers from panics and returns a 500 response.
func Recovery() bolt.MiddlewareFunc {
	return func(next bolt.Handler) bolt.Handler {
		return bolt.HandlerFunc(func(w bolt.ResponseWriter, r *bolt.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(bolt.StatusInternalServerError)
					w.Write([]byte(bolt.StatusText(bolt.StatusInternalServerError)))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
