package main

import (
	"log"
	"net/http"
)

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/app/*", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/healthz", healthzHandler)
	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}
	log.Println("Running server at 8080")
	server.ListenAndServe()
}
