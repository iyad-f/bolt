package main

import (
	"fmt"
	"io"
	"log"

	"github.com/iyad-f/bolt"
)

func main() {
	server := &bolt.Server{
		Addr: ":8080",
		Handler: bolt.HandlerFunc(func(w bolt.ResponseWriter, r *bolt.Request) {
			w.Header().Set("Content-Type", "text/plain")

			body, _ := io.ReadAll(r.Body)
			if len(body) > 0 {
				fmt.Fprintf(w, "%s %s\nBody: %s", r.Method, r.RequestURI, string(body))
			} else {
				fmt.Fprintf(w, "%s %s\nHello, World!", r.Method, r.RequestURI)
			}
		}),
	}

	fmt.Println("Server listening on :8080")
	log.Fatal(server.ListenAndServe())
}
