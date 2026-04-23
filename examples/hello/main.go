package main

import (
	"fmt"
	"io"
	"log"

	"github.com/iyad-f/bolt"
)

func main() {
	router := bolt.New()

	router.GET("/", func(w bolt.ResponseWriter, r *bolt.Request) {
		w.Write([]byte("Hello, World!"))
	})

	router.GET("/users/:id", func(w bolt.ResponseWriter, r *bolt.Request) {
		w.Write([]byte("User ID: " + r.URL.Path))
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
