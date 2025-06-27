package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"

	"github.com/google/uuid"
	"github.com/maniac-en/chirpstack/internal/auth"
	"github.com/maniac-en/chirpstack/internal/database"
	"github.com/maniac-en/chirpstack/internal/utils"
)

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
	// check for optional author_id param
	authorID := r.URL.Query().Get("author_id")
	defer r.Body.Close()
	var chirps []database.Chirp
	var err error
	if authorID != "" {
		authorUUID, err := uuid.Parse(authorID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNoContent, "invalid author id")
			return
		}
		chirps, err = cfg.DB.GetChirpsByAuthorID(r.Context(), uuid.NullUUID{
			UUID:  authorUUID,
			Valid: true,
		})
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}
	} else {
		chirps, err = cfg.DB.GetChirps(r.Context())
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}
	}
	if len(chirps) > 0 {
		// check for sort param
		sortOrder := r.URL.Query().Get("sort")
		if sortOrder == "desc" {
			sort.Slice(chirps, func(i, j int) bool { return chirps[i].CreatedAt.After(chirps[j].CreatedAt) })
		}
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

func (cfg *APIConfig) DeleteChirp(w http.ResponseWriter, r *http.Request) {
	// check for jwt existence, else return 403
	jwtToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusForbidden, "operation not allowed")
		return
	}

	// validate jwt, else return 403
	userID, err := auth.ValidateJWT(jwtToken, cfg.JWTTokenSecret)
	if err != nil {
		utils.RespondWithError(w, http.StatusForbidden, "operation not allowed")
		return
	}

	// check for chirp's existence, else return 404
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

	// check if user is authorized, else return 403
	if chirp.UserID.UUID != userID {
		utils.RespondWithError(w, http.StatusForbidden, "operation not allowed")
		return
	}

	err = cfg.DB.DeleteChirpByID(r.Context(), chirp.ID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	utils.RespondWithJSON(w, http.StatusNoContent, nil)
}

