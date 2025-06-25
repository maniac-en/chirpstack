package utils

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
)

func RemoveProfanity(chirp string) (string, bool) {
	profaneWords := []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}
	var cleanedChirpWords []string
	var profanityFound bool
	for _, word := range strings.Split(chirp, " ") {
		if len(word) == 0 {
			continue
		}
		profanityFound = false
		if slices.Contains(profaneWords, strings.ToLower(word)) {
			profanityFound = true
			cleanedChirpWords = append(cleanedChirpWords, "****")
		}
		if !profanityFound {
			cleanedChirpWords = append(cleanedChirpWords, word)
		}
	}
	cleanedChirp := strings.Join(cleanedChirpWords, " ")
	return cleanedChirp, cleanedChirp != chirp
}

func RespondWithJSON(w http.ResponseWriter, code int, payload any) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
	return nil
}

func RespondWithError(w http.ResponseWriter, code int, msg string) error {
	return RespondWithJSON(w, code, map[string]string{"error": msg})
}
