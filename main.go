package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/KrishKoria/Chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
)

type apiConfig struct {
    fileserverHits atomic.Int32
    db             *database.Queries
    platform       string
}

type User struct {
    ID        uuid.UUID `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Email     string    `json:"email"`
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
    respondWithJSON(w, code, map[string]string{"error": msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(code)
    resp, _ := json.Marshal(payload)
    w.Write(resp)
}

func cleanProfanity(content string) string {
    profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
    words := strings.Fields(content)

    for i, word := range words {
        for _, profane := range profaneWords {
            if strings.ToLower(word) == profane {
                words[i] = "****"
                break
            }
        }
    }

    return strings.Join(words, " ")
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cfg.fileserverHits.Add(1)
        next.ServeHTTP(w, r)
    })
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, `
        <html>
          <body>
            <h1>Welcome, Chirpy Admin</h1>
            <p>Chirpy has been visited %d times!</p>
          </body>
        </html>`, cfg.fileserverHits.Load())
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
    if cfg.platform != "dev" {
        w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        w.WriteHeader(http.StatusForbidden)
        w.Write([]byte("Forbidden"))
        return
    }

    err := cfg.db.DeleteAllUsers(r.Context())
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Failed to delete users")
        return
    }

    cfg.fileserverHits.Store(0)
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("All users deleted and hits reset to 0"))
}

func (cfg *apiConfig) chirpsHandler(w http.ResponseWriter, r *http.Request) {
    type chirpRequest struct {
        Body   string `json:"body"`
        UserID string `json:"user_id"`
    }

    type chirpResponse struct {
        ID        uuid.UUID `json:"id"`
        CreatedAt time.Time `json:"created_at"`
        UpdatedAt time.Time `json:"updated_at"`
        Body      string    `json:"body"`
        UserID    uuid.UUID `json:"user_id"`
    }

    var req chirpRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    if len(req.Body) > 140 {
        respondWithError(w, http.StatusBadRequest, "Chirp is too long")
        return
    }

    cleaned := cleanProfanity(req.Body)
    chirpID := uuid.New()
    createdAt := time.Now()
    updatedAt := createdAt

    params := database.CreateChirpParams{
        ID:        chirpID,
        CreatedAt: createdAt,
        UpdatedAt: updatedAt,
        Body:      cleaned,
        UserID:    uuid.MustParse(req.UserID),
    }

    chirp, err := cfg.db.CreateChirp(r.Context(), params)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Failed to create chirp")
        return
    }

    respondWithJSON(w, http.StatusCreated, chirpResponse{
        ID:        chirp.ID,
        CreatedAt: chirp.CreatedAt,
        UpdatedAt: chirp.UpdatedAt,
        Body:      chirp.Body,
        UserID:    chirp.UserID,
    })
}

func (cfg *apiConfig) usersHandler(w http.ResponseWriter, r *http.Request) {
    type userRequest struct {
        Email string `json:"email"`
    }

    var req userRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    if req.Email == "" {
        respondWithError(w, http.StatusBadRequest, "Email is required")
        return
    }

    user, err := cfg.db.CreateUser(r.Context(), req.Email)
    if err != nil {
        if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
            respondWithError(w, http.StatusConflict, "Email already exists")
            return
        }

        respondWithError(w, http.StatusInternalServerError, "Failed to create user")
        return
    }

    mappedUser := User{
        ID:        user.ID,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
        Email:     user.Email,
    }

    respondWithJSON(w, http.StatusCreated, mappedUser)
}

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
    cfg := &apiConfig{
        db:       dbQueries,
        platform: os.Getenv("PLATFORM"),
    }

    mux := http.NewServeMux()
    mux.Handle("/app/", http.StripPrefix("/app", cfg.middlewareMetricsInc(http.FileServer(http.Dir("./app")))))
    mux.HandleFunc("GET /api/healthz", readinessHandler)
    mux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
    mux.HandleFunc("POST /admin/reset", cfg.resetHandler)
    mux.HandleFunc("POST /api/users", cfg.usersHandler)
    mux.HandleFunc("POST /api/chirps", cfg.chirpsHandler)

    server := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }
    server.ListenAndServe()
}