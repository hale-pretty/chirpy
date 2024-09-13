package main

import (
	"net/http"
	"strconv"

	"github.com/hale-pretty/chirpy/database"
)

func (cfg *apiConfig) getChirpsByChirpIdHandler(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")
	chirpID, err := strconv.Atoi(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}

	dbChirp, err := cfg.DB.GetChirpByChirpId(chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp")
		return
	}

	respondWithJSON(w, http.StatusOK, database.Chirp{
		ID:       dbChirp.ID,
		AuthorID: dbChirp.AuthorID,
		Body:     dbChirp.Body,
	})
}
