package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hale-pretty/chirpy/database"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	fileserverHits int
	jwtSecret      string
}

type ChirpRequest struct {
	Body string `json:"body"`
}

type UserRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

type LoginUser struct {
	ID               int    `json:"id"`
	Email            string `json:"email"`
	JwtToken1        string `json:"token"`
	JwtRefreshToken1 string `json:"refresh_token"`
}

type AccessToken struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var db *database.DB
var jwtSecret string
var defaultExpireInSecond int

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

	// code a part to extract userID from token

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
	decoder := json.NewDecoder(r.Body)
	userRequest := UserRequest{}
	err := decoder.Decode(&userRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}
	// userRequest is a struct with data populated successfully
	userWoPW, err := db.CreateUser(userRequest.Email, userRequest.Password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
	}
	respondWithJSON(w, 201, userWoPW)
}

func loginUsersHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userRequest := UserRequest{}
	err1 := decoder.Decode(&userRequest)
	if err1 != nil {
		w.WriteHeader(http.StatusBadRequest)
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}
	// userRequest is a struct with data populated successfully
	userWoPW, ok := db.IdentifyUser(userRequest.Password)
	if !ok {
		http.Error(w, "Invalid user", http.StatusUnauthorized)
		return
	}
	var expireInSeconds = userRequest.ExpiresInSeconds
	if expireInSeconds <= 0 || expireInSeconds > defaultExpireInSecond {
		expireInSeconds = defaultExpireInSecond
	}
	// create access token
	token, err2 := createJWT(jwtSecret, userWoPW.ID, expireInSeconds)
	if err2 != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}
	// create refresh token
	userIDbyte := make([]byte, userWoPW.ID)
	_, err := rand.Read(userIDbyte)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	refreshTokenBytes := make([]byte, 32)
	rand.Read(refreshTokenBytes)
	refreshToken := hex.EncodeToString(refreshTokenBytes)
	// Save refresh token to database
	db.LoginUser(userWoPW.ID, refreshToken)

	loginUser := LoginUser{
		ID:               userWoPW.ID,
		Email:            userWoPW.Email,
		JwtToken1:        token,
		JwtRefreshToken1: refreshToken,
	}
	respondWithJSON(w, 200, loginUser)
}

func createJWT(secret string, userID, expiresInSeconds int) (string, error) {
	userIDstr := strconv.Itoa(userID)
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Second * time.Duration(expiresInSeconds))),
		Subject:   userIDstr,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func getTokenString(h http.Header) (string, error) {
	tokenString := h.Get("Authorization")
	if tokenString == "" {
		return tokenString, errors.New("authorization header not found")
	}

	// Strip off "Bearer " prefix
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	if tokenString == "" {
		return tokenString, errors.New("token not found")
	}
	return tokenString, nil
}

func updateUsersHandler(w http.ResponseWriter, r *http.Request) {
	tokenString, err1 := getTokenString(r.Header)
	if err1 != nil {
		http.Error(w, err1.Error(), http.StatusUnauthorized)
	}
	log.Println(tokenString)
	// Parse and validate the JWT
	token, err2 := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the alg is what you expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err2 != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Extract claims and userID
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
	}
	userIDStr := claims.Subject
	userID, _ := strconv.Atoi(userIDStr)

	// Get the body to update
	decoder := json.NewDecoder(r.Body)
	userRequest := UserRequest{}
	err3 := decoder.Decode(&userRequest)
	if err3 != nil {
		w.WriteHeader(http.StatusBadRequest)
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	resp, ok := db.UpdateUser(userID, userRequest.Email, userRequest.Password)
	if !ok {
		http.Error(w, "User is not found", http.StatusUnauthorized)
	}
	respondWithJSON(w, 200, resp)
}

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	tokenString, err1 := getTokenString(r.Header)
	if err1 != nil {
		http.Error(w, err1.Error(), http.StatusUnauthorized)
		return
	}
	userID, refreshTokenOk := db.RefreshNewAccessToken(tokenString)
	if !refreshTokenOk {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	secondAccessToken, err2 := createJWT(jwtSecret, userID, defaultExpireInSecond)
	if err2 != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}
	resp := AccessToken{
		Token: secondAccessToken,
	}
	respondWithJSON(w, 200, resp)
}

func revokeHandler(w http.ResponseWriter, r *http.Request) {
	tokenString, err1 := getTokenString(r.Header)
	if err1 != nil {
		http.Error(w, err1.Error(), http.StatusUnauthorized)
		return
	}
	revokeTokenOk := db.RevokeRefreshToken(tokenString)
	if !revokeTokenOk {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(204)
}

func main() {
	// create database
	var err1 error
	db, err1 = database.NewDB("database.json")
	if err1 != nil {
		log.Fatalf("Failed to initialize database: %v", err1)
	}

	// load JWT_SECRET
	err2 := godotenv.Load()
	if err2 != nil {
		log.Fatalf("Error loading .env file")
	}
	jwtSecret = os.Getenv("JWT_SECRET")

	// set default expiration time for access token
	defaultExpireInSecond = 3600

	// create mux
	mux := http.NewServeMux()
	apiCfg := &apiConfig{
		fileserverHits: 0,
		jwtSecret:      jwtSecret,
	}
	fileServer := http.FileServer(http.Dir("."))

	mux.Handle("/app/*", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(fileServer)))
	mux.HandleFunc("GET /admin/metrics", apiCfg.hitsHandler)
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("/api/reset", apiCfg.resetHandler)
	mux.HandleFunc("POST /api/chirps", createChirpHandler)
	mux.HandleFunc("GET /api/chirps", getFullChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{id}", getChirpHandler)
	mux.HandleFunc("POST /api/users", createUsersHandler)
	mux.HandleFunc("POST /api/login", loginUsersHandler)
	mux.HandleFunc("PUT /api/users", updateUsersHandler)
	mux.HandleFunc("POST /api/refresh", refreshHandler)
	mux.HandleFunc("POST /api/revoke", revokeHandler)
	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}
	log.Println("Running server at 8080")
	server.ListenAndServe()
}
