package main

import (
	"log"
	"net/http"
	"os"

	"github.com/hale-pretty/chirpy/database"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	fileserverHits int
	DB             *database.DB
	jwtSecret      string
	polkaAPIKey    string
}

var defaultExpireInSecond int

func main() {
	// create database
	db, err := database.NewDB("database.json")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// load JWT_SECRET
	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	// load POLKA_API_KEY
	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	polkaAPIKey := os.Getenv("POLKA_API_KEY")
	if polkaAPIKey == "" {
		log.Fatal("POLKA_KEY environment variable is not set")
	}
	// set default expiration time for access token
	defaultExpireInSecond = 3600

	// create mux
	mux := http.NewServeMux()
	apiCfg := apiConfig{
		fileserverHits: 0,
		DB:             db,
		jwtSecret:      jwtSecret,
		polkaAPIKey:    polkaAPIKey,
	}
	fileServer := http.FileServer(http.Dir("."))

	mux.Handle("/app/*", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(fileServer)))
	mux.HandleFunc("GET /admin/metrics", apiCfg.hitsHandler)
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("/api/reset", apiCfg.resetHandler)
	mux.HandleFunc("POST /api/chirps", apiCfg.createChirpHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirpsByChirpIdHandler)
	mux.HandleFunc("POST /api/users", apiCfg.createUsersHandler)
	mux.HandleFunc("POST /api/login", apiCfg.loginUsersHandler)
	mux.HandleFunc("PUT /api/users", apiCfg.updateUsersHandler)
	mux.HandleFunc("POST /api/refresh", apiCfg.refreshHandler)
	mux.HandleFunc("POST /api/revoke", apiCfg.revokeHandler)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.deleteChirpHandler)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.polkaWebhooksHandler)
	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}
	log.Println("Running server at 8080")
	server.ListenAndServe()
}
