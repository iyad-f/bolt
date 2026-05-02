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
	"slices"
	"strconv"
	"strings"

	"github.com/iyad-f/bolt"
)

// CORSConfig defines the configuration for the CORS middleware.
type CORSConfig struct {
	// AllowedOrigins is a list of origins that may access the resource.
	// Supports wildcard patterns like "http://*.example.com". Defaults to ["*"].
	AllowedOrigins []string

	// AllowedMethods is a list of methods the client is allowed to use. Defaults to GET, POST, HEAD.
	AllowedMethods []string

	// AllowedHeaders is a list of headers the client is allowed to send.
	AllowedHeaders []string

	// ExposedHeaders is a list of headers that are safe to expose to the browser.
	ExposedHeaders []string

	// AllowCredentials indicates whether cookies and credentials are allowed.
	AllowCredentials bool

	// MaxAge is the duration in seconds that preflight responses are cached by the browser.
	MaxAge int

	// AllowOriginFunc is a custom function to validate the origin. It is called
	// when no match is found in AllowedOrigins.
	AllowOriginFunc func(origin string) bool
}

// DefaultCORSConfig returns a CORSConfig that allows all origins with GET, POST, and HEAD methods.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "HEAD"},
	}
}

type wildcardOrigin struct {
	prefix string
	suffix string
}

// CORS returns a middleware that handles Cross-Origin Resource Sharing.
func CORS(config CORSConfig) bolt.MiddlewareFunc {
	if len(config.AllowedOrigins) == 0 {
		config.AllowedOrigins = []string{"*"}
	}

	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = []string{"GET", "POST", "HEAD"}
	}

	allowAllOrigins := slices.Contains(config.AllowedOrigins, "*")
	if allowAllOrigins && config.AllowCredentials {
		panic("bolt: CORS misconfigured: wildcard origin with credentials is insecure")
	}

	for idx, header := range config.AllowedHeaders {
		config.AllowedHeaders[idx] = strings.ToLower(header)
	}

	allowMethodsStr := strings.Join(config.AllowedMethods, ", ")
	allowHeadersStr := strings.Join(config.AllowedHeaders, ", ")
	exposeHeadersStr := strings.Join(config.ExposedHeaders, ", ")
	maxAgeStr := strconv.Itoa(config.MaxAge)

	var wildcardOrigins []wildcardOrigin
	for _, origin := range config.AllowedOrigins {
		if origin == "*" {
			continue
		}
		if before, after, found := strings.Cut(origin, "*"); found {
			wildcardOrigins = append(wildcardOrigins, wildcardOrigin{
				prefix: before,
				suffix: after,
			})
		}
	}

	return func(next bolt.Handler) bolt.Handler {
		return bolt.HandlerFunc(func(w bolt.ResponseWriter, r *bolt.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Add("Vary", "Origin")

			if !isOriginAllowed(origin, config, wildcardOrigins, allowAllOrigins) {
				next.ServeHTTP(w, r)
				return
			}

			if allowAllOrigins {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Preflight request
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				w.Header().Add("Vary", "Access-Control-Request-Method")
				w.Header().Add("Vary", "Access-Control-Request-Headers")
				w.Header().Set("Access-Control-Allow-Methods", allowMethodsStr)
				w.Header().Set("Access-Control-Allow-Headers", allowHeadersStr)

				if config.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", maxAgeStr)
				}

				w.WriteHeader(bolt.StatusNoContent)
				return
			}

			// Actual request
			if exposeHeadersStr != "" {
				w.Header().Set("Access-Control-Expose-Headers", exposeHeadersStr)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isOriginAllowed(origin string, config CORSConfig, wildcardOrigins []wildcardOrigin, allowAllOrigins bool) bool {
	if allowAllOrigins {
		return true
	}

	if slices.Contains(config.AllowedOrigins, origin) {
		return true
	}

	for _, wildcard := range wildcardOrigins {
		if strings.HasPrefix(origin, wildcard.prefix) &&
			strings.HasSuffix(origin, wildcard.suffix) &&
			len(origin) > len(wildcard.prefix)+len(wildcard.suffix) {
			return true
		}
	}

	if config.AllowOriginFunc != nil {
		return config.AllowOriginFunc(origin)
	}

	return false
}
