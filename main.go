// Package chirpstack is a learning-project mimicking the backend stack of twitter
package main

import (
	"log"
	"net/http"

	"github.com/maniac-en/chirpstack/internal/cfg"
	"github.com/maniac-en/chirpstack/internal/server"
)

func main() {
	var apiCfg cfg.APIConfig
	mux := http.NewServeMux()
	fileserverHandler := http.StripPrefix("/app", http.FileServer(http.Dir('.')))

	// app
	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(fileserverHandler))

	// api
	mux.HandleFunc("GET /api/healthz", server.HealthzHandler)
	mux.HandleFunc("POST /api/validate_chirp", server.ValidateChirpHandler)

	// admin
	mux.HandleFunc("GET /admin/metrics", apiCfg.MetricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.ResetMetricsHandler)

	loggedMux := server.LogMiddleware(mux)
	log.Fatal(http.ListenAndServe(":8080", loggedMux))
}
