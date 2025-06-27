package api

import (
	"encoding/json"
	"io"
	"net/mail"
	"net/http"

	"github.com/maniac-en/chirpstack/internal/auth"
	"github.com/maniac-en/chirpstack/internal/database"
	"github.com/maniac-en/chirpstack/internal/utils"
)

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

func (cfg *APIConfig) UpdateUser(w http.ResponseWriter, r *http.Request) {
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
		Email    string `json:"email"`
		Password string `json:"password"`
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

	updateUserParams := database.UpdateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
		ID:             userID,
	}

	updatedUserInfo, err := cfg.DB.UpdateUser(r.Context(), updateUserParams)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, updatedUserInfo)
}