// Package api provides the handlers/middlewares
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/mail"
	"sync/atomic"

	"github.com/maniac-en/chirpstack/internal/database"
	"github.com/maniac-en/chirpstack/internal/utils"
)

type Platform string

const (
	PlatformDev  Platform = "dev"
	PlatformProd Platform = "prod"
)

func (p Platform) IsValid() bool {
	switch p {
	case PlatformDev, PlatformProd:
		return true
	default:
		return false
	}
}

func ValidPlatforms() []Platform {
	return []Platform{PlatformDev, PlatformProd}
}

func ParsePlatform(s string) (Platform, error) {
	p := Platform(s)
	if !p.IsValid() {
		return "", fmt.Errorf("invalid platform '%s', must be one of: %v", s, ValidPlatforms())
	}
	return p, nil
}

type APIConfig struct {
	fileserverHits atomic.Int32
	DB             *database.Queries
	PLATFORM       Platform
}

func (cfg *APIConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		w.Header().Add("Cache-Control", "no-cache")
		next.ServeHTTP(w, r)
	})
}

func (cfg *APIConfig) LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (cfg *APIConfig) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `
		<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`, cfg.fileserverHits.Load())
}

func (cfg *APIConfig) ResetHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.PLATFORM != PlatformDev {
		utils.RespondWithError(w, http.StatusForbidden, "Operation not allowed")
		return
	}
	cfg.fileserverHits.Store(0)
	err := cfg.DB.TruncateUsers(r.Context())
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte(http.StatusText(http.StatusNoContent)))
}

func (cfg *APIConfig) HealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *APIConfig) ValidateChirpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	type requestBody struct {
		Body string `json:"body"`
	}
	type responseBody struct {
		Valid       bool   `json:"valid,omitempty"`
		CleanedBody string `json:"cleaned_body,omitempty"`
	}
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	params := requestBody{}
	if err := json.Unmarshal(data, &params); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if len(params.Body) > 140 {
		utils.RespondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	if cleanedChirp, cleaned := utils.RemoveProfanity(params.Body); cleaned {
		utils.RespondWithJSON(w, http.StatusOK, responseBody{
			CleanedBody: cleanedChirp,
		})
	} else {
		utils.RespondWithJSON(w, http.StatusOK, responseBody{
			Valid: true,
		})
	}
}

func (cfg *APIConfig) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	type requestBody struct {
		Email string `json:"email"`
	}
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	params := requestBody{}
	if err := json.Unmarshal(data, &params); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	_, err = mail.ParseAddress(params.Email)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid email address")
		return
	}

	usr, err := cfg.DB.CreateUser(r.Context(), params.Email)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	utils.RespondWithJSON(w, http.StatusCreated, usr)
}

func (cfg *APIConfig) CreateChirps(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	type requestBody database.CreateChirpParams
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	params := requestBody{}
	if err := json.Unmarshal(data, &params); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if len(params.Body) > 140 {
		utils.RespondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	if cleanedChirp, cleaned := utils.RemoveProfanity(params.Body); cleaned {
		params.Body = cleanedChirp
	}

	chirp, err := cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams(params))
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	utils.RespondWithJSON(w, http.StatusCreated, chirp)
}

func (cfg *APIConfig) GetChirps(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	defer r.Body.Close()
	chirps, err := cfg.DB.GetChirps(r.Context())
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	if len(chirps) > 0 {
		utils.RespondWithJSON(w, http.StatusOK, chirps)
	} else {
		utils.RespondWithJSON(w, http.StatusOK, []database.Chirp{})
	}
}
