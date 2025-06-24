// Package chirpstack is a learning-project mimicking the backend stack of twitter
package main

import (
	"log"
	"net/http"
)

func main() {
	serveMuxHandler := http.NewServeMux()
	serveMuxHandler.Handle("/", http.FileServer(http.Dir('.')))
	server := http.Server{
		Addr:    ":8080",
		Handler: serveMuxHandler,
	}
	log.Fatal(server.ListenAndServe())
}
