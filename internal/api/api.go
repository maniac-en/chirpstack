// Package api provides the handlers/middlewares
package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/mail"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/maniac-en/chirpstack/internal/auth"
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
	Platform       Platform
	JWTTokenSecret string
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
	if cfg.Platform != PlatformDev {
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
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *APIConfig) HealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *APIConfig) ValidateChirpHandler(w http.ResponseWriter, r *http.Request) {
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
	type requestBody struct {
		Password string `json:"password"`
		Email    string `json:"email"`
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

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		if err.Error() == auth.ErrPasswordTooLong {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}
	}

	newUserParams := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	}

	newUser, err := cfg.DB.CreateUser(r.Context(), newUserParams)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	utils.RespondWithJSON(w, http.StatusCreated, newUser)
}

func (cfg *APIConfig) LoginUser(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Password        string `json:"password"`
		Email           string `json:"email"`
		ExpiryInSeconds int    `json:"expiry_in_seconds,omitempty"`
	}
	type responseBody struct {
		database.User
		Token string `json:"token"`
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

	// check if passed email is valid or not
	_, err = mail.ParseAddress(params.Email)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid email address")
		return
	}

	// fetch user info from DB
	storedUserInfo, err := cfg.DB.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	// validate user's password hash
	err = auth.CheckPasswordHash(params.Password, storedUserInfo.HashedPassword)
	if err != nil {
		if err.Error() == auth.ErrIncorrectEmailOrPassword {
			utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
			return
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}
	}

	// check/set expiry_in_seconds
	var expiryDuration time.Duration
	maxExpiry := time.Hour
	if params.ExpiryInSeconds == 0 || time.Duration(params.ExpiryInSeconds)*time.Second > maxExpiry {
		expiryDuration = maxExpiry
	} else {
		expiryDuration = time.Duration(params.ExpiryInSeconds) * time.Second
	}

	jwtToken, err := auth.MakeJWT(storedUserInfo.ID, cfg.JWTTokenSecret, expiryDuration)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	res := responseBody{
		User:  storedUserInfo,
		Token: jwtToken,
	}
	utils.RespondWithJSON(w, http.StatusOK, res)
}

func (cfg *APIConfig) CreateChirps(w http.ResponseWriter, r *http.Request) {
	jwtToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userID, err := auth.ValidateJWT(jwtToken, cfg.JWTTokenSecret)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	type requestBody struct {
		Body string `json:"body"`
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
		params.Body = cleanedChirp
	}

	chirp, err := cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		Body: params.Body,
		UserID: uuid.NullUUID{
			UUID:  userID,
			Valid: true,
		},
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	utils.RespondWithJSON(w, http.StatusCreated, chirp)
}

func (cfg *APIConfig) GetChirps(w http.ResponseWriter, r *http.Request) {
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

func (cfg *APIConfig) GetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("id")
	if chirpID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "No chirp ID passed")
		return
	}
	chirpUUID, err := uuid.Parse(chirpID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Invalid chirp ID passed")
		return
	}

	chirp, err := cfg.DB.GetChirpByID(r.Context(), chirpUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.RespondWithError(w, http.StatusNotFound, "No chirp found")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, chirp)
}
