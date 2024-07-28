package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/hale-pretty/chirpy/database"
)

type apiConfig struct {
	fileserverHits int
}

type ChirpRequest struct {
	Body string `json:"body"`
}

type UserRequest struct {
	Email string `json:"email"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var db *database.DB

// Handler for /healthz

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) hitsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	html := fmt.Sprintf(`
		<html>

		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>

		</html>`, cfg.fileserverHits)
	w.Write([]byte(html))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
}

// Handler for /admin/metrics

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Handler for POST /api/chirps
func respondWithError(w http.ResponseWriter, code int, msg string) {
	errResp := ErrorResponse{Error: msg}
	respondWithJSON(w, code, errResp)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Something went wrong"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
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

func createChirpHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	chirpRequest := ChirpRequest{}
	err := decoder.Decode(&chirpRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}
	// chirpRequest is a struct with data populated successfully

	if len(chirpRequest.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	cleanedBody := cleanProfanity(chirpRequest.Body)
	chirp, err := db.CreateChirp(cleanedBody)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
	}
	respondWithJSON(w, 201, chirp)
}

func getFullChirpsHandler(w http.ResponseWriter, r *http.Request) {
	chirpsMap := db.Data.Chirps
	resp := make([]database.Chirp, 0, len(chirpsMap))
	for _, chirp := range chirpsMap {
		resp = append(resp, chirp)
	}
	respondWithJSON(w, 200, resp)
}

func getChirpHandler(w http.ResponseWriter, r *http.Request) {
	chirpIdStr := r.PathValue("id")
	chirpIdInt, err := strconv.Atoi(chirpIdStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid path")
	}
	chirpsMap := db.Data.Chirps
	respChirp, existedId := chirpsMap[chirpIdInt]
	if !existedId {
		http.Error(w, "404 page not found", http.StatusNotFound)
	}
	respondWithJSON(w, 200, respChirp)
}

func createUsersHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Body)
	decoder := json.NewDecoder(r.Body)
	userRequest := UserRequest{}
	err := decoder.Decode(&userRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}
	log.Println(userRequest)
	// userRequest is a struct with data populated successfully
	log.Println(userRequest)
	user, err := db.CreateUser(userRequest.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
	}
	respondWithJSON(w, 201, user)
}

func main() {
	var err error
	db, err = database.NewDB("database.json")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	mux := http.NewServeMux()
	apiCfg := &apiConfig{fileserverHits: 0}
	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/app/*", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(fileServer)))
	mux.HandleFunc("GET /admin/metrics", apiCfg.hitsHandler)
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("/api/reset", apiCfg.resetHandler)
	mux.HandleFunc("POST /api/chirps", createChirpHandler)
	mux.HandleFunc("GET /api/chirps", getFullChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{id}", getChirpHandler)
	mux.HandleFunc("POST /api/users", createUsersHandler)
	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}
	log.Println("Running server at 8080")
	server.ListenAndServe()
}
