// Package chirpstack is a learning-project mimicking the backend stack of twitter
package main

import "net/http"

func main() {
	serveMuxHandler := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: serveMuxHandler,
	}
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
