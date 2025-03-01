package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"encoding/json"
	"strings"
    _ "github.com/lib/pq"
    "os"
    "database/sql"
    "github.com/KrishKoria/Chirpy/internal/database"
)


type apiConfig struct {
	fileserverHits atomic.Int32
    db *database.Queries
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
	w.Header().Set("Content-Type", "text/plain; charset=utf-8");
	w.WriteHeader(http.StatusOK);
	w.Write([]byte("OK"));
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w,`
        <html>
          <body>
            <h1>Welcome, Chirpy Admin</h1>
            <p>Chirpy has been visited %d times!</p>
          </body>
        </html>`, cfg.fileserverHits.Load())
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
    cfg.fileserverHits.Store(0)
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Hits reset to 0"))
}

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type chirpRequest struct {
		Body string `json:"body"`
	}
	
	type chirpResponse struct {
		Cleaned_Body string `json:"cleaned_body"`
	}
	var req chirpRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
        return
    }

    if len(req.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
        return
    }

    cleaned := cleanProfanity(req.Body)
	respondWithJSON(w, http.StatusOK, chirpResponse{Cleaned_Body: cleaned})
}
func main()  {
    
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
        db: dbQueries,
    }
    
    mux := http.NewServeMux();
	mux.Handle("/app/", http.StripPrefix("/app", cfg.middlewareMetricsInc(http.FileServer(http.Dir("./app")))));
	mux.HandleFunc("GET /api/healthz", readinessHandler);
	mux.HandleFunc("GET /admin/metrics", cfg.metricsHandler);
	mux.HandleFunc("POST /admin/reset", cfg.resetHandler);
	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler);
	server:= &http.Server{
		Addr: ":8080",
		Handler: mux,	
	}
	server.ListenAndServe();
}