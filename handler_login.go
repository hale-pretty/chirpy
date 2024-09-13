package main

import (
	"encoding/json"
	"net/http"

	"github.com/hale-pretty/chirpy/internal/auth"
)

type LoginUser struct {
	ID               int    `json:"id"`
	Email            string `json:"email"`
	JwtToken1        string `json:"token"`
	JwtRefreshToken1 string `json:"refresh_token"`
	IsChirpyRed      bool   `json:"is_chirpy_red"`
}

func (cfg *apiConfig) loginUsersHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userRequest := UserRequest{}
	err := decoder.Decode(&userRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}
	// userRequest is a struct with data populated successfully
	userWoPW, ok := cfg.DB.IdentifyUser(userRequest.Password)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Invalid user")
		return
	}
	var expireInSeconds = userRequest.ExpiresInSeconds
	if expireInSeconds <= 0 || expireInSeconds > defaultExpireInSecond {
		expireInSeconds = defaultExpireInSecond
	}
	// create access token
	token, err := auth.CreateJWT(cfg.jwtSecret, userWoPW.ID, expireInSeconds)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating token")
		return
	}
	// create refresh token
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create refresh token")
		return
	}
	// Save refresh token to database
	cfg.DB.LoginUser(userWoPW.ID, refreshToken)

	loginUser := LoginUser{
		ID:               userWoPW.ID,
		Email:            userWoPW.Email,
		JwtToken1:        token,
		JwtRefreshToken1: refreshToken,
		IsChirpyRed:      userWoPW.IsChirpyRed,
	}
	respondWithJSON(w, 200, loginUser)
}
