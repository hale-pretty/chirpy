package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hale-pretty/chirpy/database"
	"github.com/hale-pretty/chirpy/internal/auth"
)

type ChirpyRedRequest struct {
	Event string       `json:"event"`
	Data  *DataWebhook `json:"data"`
}

type DataWebhook struct {
	UserID int `json:"user_id"`
}

func (cfg *apiConfig) polkaWebhooksHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Check API Key
	APIKey, err := auth.GetAPIkey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Cannot find api key")
		return
	}
	if APIKey != cfg.polkaAPIKey {
		respondWithError(w, http.StatusUnauthorized, "API key is invalid")
		return
	}

	// 2. Decode Request Body
	chirpyRedDataWebhook := &DataWebhook{}
	decoder := json.NewDecoder(r.Body)
	chirpyRedRequest := ChirpyRedRequest{
		Data: chirpyRedDataWebhook,
	}
	err = decoder.Decode(&chirpyRedRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	// 2. Check the event
	if chirpyRedRequest.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, "We don't care about any other events")
		return
	}

	// 3. Save to database
	err = cfg.DB.IsChirpyRed(chirpyRedRequest.Data.UserID)
	if err != nil {
		if errors.Is(err, database.ErrNotExist) {
			respondWithError(w, http.StatusNotFound, "Couldn't find user")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
