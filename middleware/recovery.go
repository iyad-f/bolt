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
