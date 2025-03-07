package main

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/KrishKoria/Chirpy/internal/database"
	"github.com/joho/godotenv"
)


func main() {
    err := godotenv.Load()
    if err != nil {
        panic("Error loading .env file")
    }

    dbURL := os.Getenv("DB_URL")
    if dbURL == "" {
        panic("DB_URL environment variable is not set")
    }
    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        panic(err)
    }
    defer db.Close()

    dbQueries := database.New(db)
    cfg := &APIConfig{
        DB:       dbQueries,
        Platform: os.Getenv("PLATFORM"),
        JWTSecret: os.Getenv("JWT_SECRET"),
    }

    mux := http.NewServeMux()
    mux.Handle("/app/", http.StripPrefix("/app", cfg.middlewareMetricsInc(http.FileServer(http.Dir("./app")))))
    mux.HandleFunc("GET /api/healthz", ReadinessHandler)
    mux.HandleFunc("GET /admin/metrics", cfg.MetricsHandler)
    mux.HandleFunc("GET /api/chirps", cfg.getAllChirpsHandler)
    mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.getChirpHandler)
    mux.HandleFunc("POST /admin/reset", cfg.ResetHandler)
    mux.HandleFunc("POST /api/users", cfg.UsersHandler)
    mux.HandleFunc("POST /api/chirps", cfg.chirpsHandler)
    mux.HandleFunc("POST /api/login", cfg.loginHandler)
    mux.HandleFunc("POST /api/refresh", cfg.refreshHandler)
    mux.HandleFunc("POST /api/revoke", cfg.revokeHandler)
    server := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }
    server.ListenAndServe()
}