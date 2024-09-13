package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/hale-pretty/chirpy/internal/auth"
)

type ChirpRequest struct {
	Body string `json:"body"`
}

func cleanProfanity(msg string) string {
	processedMsgSlices := strings.Split(msg, " ")
	for idx, word := range processedMsgSlices {
		lowerWord := strings.ToLower(word)
		if lowerWord == "kerfuffle" || lowerWord == "sharbert" || lowerWord == "fornax" {
			processedMsgSlices[idx] = "****"
		}
	}
	return strings.Join(processedMsgSlices, " ")
}

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get the Author ID
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Cannot find JWT: %v", err))
		return
	}
	// Parse and validate the JWT
	userIDStr, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Cannot validate JWT: %v", err))
		return
	}
	userID, _ := strconv.Atoi(userIDStr)

	// 2. Decode Request Body
	decoder := json.NewDecoder(r.Body)
	chirpRequest := ChirpRequest{}
	err = decoder.Decode(&chirpRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	// 3. Validate chirp body
	if len(chirpRequest.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	cleanedBody := cleanProfanity(chirpRequest.Body)

	// 4. Create Chirp
	chirp, err := cfg.DB.CreateChirp(cleanedBody, userID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
	}
	respondWithJSON(w, 201, chirp)
}
