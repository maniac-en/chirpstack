package api

import (
	"encoding/json"
	"io"
	"net/mail"
	"net/http"

	"github.com/google/uuid"
	"github.com/maniac-en/chirpstack/internal/auth"
	"github.com/maniac-en/chirpstack/internal/database"
	"github.com/maniac-en/chirpstack/internal/utils"
)

func (cfg *APIConfig) LoginUser(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type responseBody struct {
		database.User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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

	// generate an access token for user (jwt)
	jwtToken, err := auth.MakeJWT(storedUserInfo.ID, cfg.JWTTokenSecret)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	// generate and store a refresh token for user
	refreshTokenString, _ := auth.MakeRefreshToken()
	refreshToken, err := cfg.DB.StoreRefreshToken(r.Context(), database.StoreRefreshTokenParams{
		Token: refreshTokenString,
		UserID: uuid.NullUUID{
			UUID:  storedUserInfo.ID,
			Valid: true,
		},
	})
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	res := responseBody{
		User:         storedUserInfo,
		Token:        jwtToken,
		RefreshToken: refreshToken.Token,
	}
	utils.RespondWithJSON(w, http.StatusOK, res)
}

func (cfg *APIConfig) RefreshUserToken(w http.ResponseWriter, r *http.Request) {
	type responseBody struct {
		Token string `json:"token"`
	}
	// get refresh token from headers
	refreshTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// check if it's a valid refresh token, and if yes, return it,
	// if not found or valid, return 401
	_, err = cfg.DB.GetUserFromRefreshToken(r.Context(), refreshTokenString)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, responseBody{refreshTokenString})
}

func (cfg *APIConfig) RevokeUserToken(w http.ResponseWriter, r *http.Request) {
	// get refresh token from headers
	refreshTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	_, err = cfg.DB.RevokeRefreshToken(r.Context(), refreshTokenString)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	utils.RespondWithJSON(w, http.StatusNoContent, nil)
}