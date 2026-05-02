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

package main

import (
	"fmt"
	"io"
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

	router.GET("/panic", func(w bolt.ResponseWriter, r *bolt.Request) {
		panic("something went wrong!")
	})

	router.GET("/users/:id", func(w bolt.ResponseWriter, r *bolt.Request) {
		w.Write([]byte("User ID: " + r.PathValue("id")))
	})

	router.POST("/echo", func(w bolt.ResponseWriter, r *bolt.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Fprintf(w, "Echo: %s", string(body))
	})

	server := &bolt.Server{
		Addr:    ":8080",
		Handler: router,
	}

	fmt.Println("Server listening on :8080")
	log.Fatal(server.ListenAndServe())
}
