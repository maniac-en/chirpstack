// Package chirpstack is a learning-project mimicking the backend stack of twitter
package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/maniac-en/chirpstack/internal/api"
	"github.com/maniac-en/chirpstack/internal/database"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err1 := sql.Open("postgres", os.Getenv("DB_URL"))
	platform, err2 := api.ParsePlatform(os.Getenv("PLATFORM"))
	err = errors.Join(err1, err2)
	if err != nil {
		log.Fatal(err)
	}

	jwtTokenSecret := os.Getenv("JWT_TOKEN_SECRET")
	if jwtTokenSecret == "" {
		log.Fatal("empty secret found for JWT signing")
	}

	polkaAPIKey := os.Getenv("POLKA_KEY")

	apiCfg := api.APIConfig{
		DB:             database.New(db),
		Platform:       platform,
		JWTTokenSecret: jwtTokenSecret,
		PolkaAPIKey:    polkaAPIKey,
	}
	mux := http.NewServeMux()
	fileserverHandler := http.StripPrefix("/app", http.FileServer(http.Dir('.')))

	// app
	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(fileserverHandler))

	// api
	mux.HandleFunc("GET /api/health", apiCfg.HealthHandler)

	mux.HandleFunc("GET /api/chirps", apiCfg.GetChirps)
	mux.HandleFunc("GET /api/chirps/{id}", apiCfg.GetChirpByID)

	mux.HandleFunc("POST /api/chirps", apiCfg.CreateChirps)
	mux.HandleFunc("POST /api/users", apiCfg.CreateUser)
	mux.HandleFunc("POST /api/login", apiCfg.LoginUser)
	mux.HandleFunc("POST /api/refresh", apiCfg.RefreshUserToken)
	mux.HandleFunc("POST /api/revoke", apiCfg.RevokeUserToken)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.UpgradeUser)

	mux.HandleFunc("PUT /api/users", apiCfg.UpdateUser)

	mux.HandleFunc("DELETE /api/chirps/{id}", apiCfg.DeleteChirp)

	// admin
	mux.HandleFunc("GET /admin/metrics", apiCfg.MetricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.ResetHandler)

	loggedMux := apiCfg.LogMiddleware(mux)
	log.Fatal(http.ListenAndServe(":8080", loggedMux))
}
