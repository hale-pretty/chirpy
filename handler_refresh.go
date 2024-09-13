package main

import (
	"fmt"
	"net/http"

	"github.com/hale-pretty/chirpy/internal/auth"
)

type AccessToken struct {
	Token string `json:"token"`
}

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Cannot find JWT: %v", err))
		return
	}

	userID, refreshTokenOk := cfg.DB.RefreshNewAccessToken(tokenString)
	if !refreshTokenOk {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	secondAccessToken, err := auth.CreateJWT(cfg.jwtSecret, userID, defaultExpireInSecond)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating token")
		return
	}

	resp := AccessToken{
		Token: secondAccessToken,
	}
	respondWithJSON(w, 200, resp)
}

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Cannot find JWT: %v", err))
		return
	}

	revokeTokenOk := cfg.DB.RevokeRefreshToken(tokenString)
	if !revokeTokenOk {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
