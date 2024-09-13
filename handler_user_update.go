package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hale-pretty/chirpy/internal/auth"
)

func (cfg *apiConfig) updateUsersHandler(w http.ResponseWriter, r *http.Request) {
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

	// 2. Get the body to update
	decoder := json.NewDecoder(r.Body)
	userRequest := UserRequest{}
	err3 := decoder.Decode(&userRequest)
	if err3 != nil {
		w.WriteHeader(http.StatusBadRequest)
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	// 3. Update info to database
	resp, ok := cfg.DB.UpdateUser(userID, userRequest.Email, userRequest.Password)
	if !ok {
		http.Error(w, "User is not found", http.StatusUnauthorized)
	}
	respondWithJSON(w, 200, resp)
}
