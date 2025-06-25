// Package chirpstack is a learning-project mimicking the backend stack of twitter
package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/maniac-en/chirpstack/internal/cfg"
	"github.com/maniac-en/chirpstack/internal/database"
	"github.com/maniac-en/chirpstack/internal/server"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	// dbQueries := database.New(db)
	//
	// var apiCfg cfg.APIConfig
	apiCfg := cfg.APIConfig{
		DBQueries: database.New(db),
	}
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
