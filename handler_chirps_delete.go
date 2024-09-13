package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/hale-pretty/chirpy/internal/auth"
)

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Check authorization
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

	// 2. Get chirp id
	chirpIdStr := r.PathValue("chirpID")
	chirpIdInt, err := strconv.Atoi(chirpIdStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid path: %v", err))
		return
	}

	// 3. Delete chirp in the database
	err = cfg.DB.DeleteChirp(userID, chirpIdInt)
	if err != nil {
		respondWithError(w, http.StatusForbidden, fmt.Sprintf("Cannot delete this chirp: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
