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
