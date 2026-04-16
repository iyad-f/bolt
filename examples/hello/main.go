package main

import (
	"fmt"
	"log"

	"github.com/iyad-f/bolt"
)

func main() {
	server := &bolt.Server{
		Addr: ":8080",
		Handler: bolt.HandlerFunc(func(w bolt.ResponseWriter, r *bolt.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(bolt.StatusOK)
			w.Write([]byte("Hello, World!"))
		}),
	}

	fmt.Println("Server listening on :8080")
	log.Fatal(server.ListenAndServe())
}
