# Bolt

An HTTP/1.1 server built from raw TCP sockets in Go. No `net/http` dependency for the core server -- everything from request parsing to response writing is implemented from scratch.

This is a learning project. I wanted to understand what actually happens under the hood when an HTTP server handles a request. It's not meant to be a production library, though it might evolve into something more serious down the line.

## What it does

- Parses HTTP/1.1 requests (method, headers, body, chunked transfer encoding)
- Writes well-formed HTTP responses with proper status codes and headers
- Radix tree router with path parameters (`:id`) and catch-all (`*filepath`) support
- Keep-alive connections with configurable timeouts (read, write, idle)
- TLS/HTTPS support
- Static file serving with ETag caching, range requests, and directory traversal protection
- Middleware system with chaining
- Graceful shutdown

## Built-in middleware

- **Logger** -- logs method, path, status, and duration
- **Recovery** -- catches panics and returns 500
- **Compress** -- gzip compression
- **CORS** -- configurable origin/method/header policies, preflight handling
- **RateLimit** -- sliding window rate limiting with pluggable store backend

## Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/iyad-f/bolt"
	"github.com/iyad-f/bolt/middleware"
)

func main() {
	router := bolt.New()

	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())

	router.GET("/", func(w bolt.ResponseWriter, r *bolt.Request) {
		w.Write([]byte("Hello, World!"))
	})

	router.GET("/users/:id", func(w bolt.ResponseWriter, r *bolt.Request) {
		w.Write([]byte("User ID: " + r.PathValue("id")))
	})

	server := &bolt.Server{
		Addr:    ":8080",
		Handler: router,
	}

	fmt.Println("Server listening on :8080")
	log.Fatal(server.ListenAndServe())
}
```

## Running tests

```
go test ./...
```

## Acknowledgements

The radix tree implementation is based on [httprouter](https://github.com/julienschmidt/httprouter) by Julien Schmidt (BSD license).

## License

Apache 2.0
