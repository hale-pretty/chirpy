package main

import (
	"encoding/json"
	"net/http"
)

type UserRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

func (cfg *apiConfig) createUsersHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userRequest := UserRequest{}
	err := decoder.Decode(&userRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}
	// userRequest is a struct with data populated successfully
	userWoPW, err := cfg.DB.CreateUser(userRequest.Email, userRequest.Password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
	}
	respondWithJSON(w, 201, userWoPW)
}
