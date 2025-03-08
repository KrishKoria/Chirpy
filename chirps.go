package main

import (
	"encoding/json"
	"net/http"
	"time"
    "database/sql"
	"github.com/KrishKoria/Chirpy/internal/database"
	"github.com/google/uuid"
    "github.com/KrishKoria/Chirpy/internal/auth"
    "sort"
)


func (cfg *APIConfig) chirpsHandler(w http.ResponseWriter, r *http.Request) {
    type chirpRequest struct {
        Body   string `json:"body"`
    }

    tokenString, err := auth.GetBearerToken(r.Header)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Authentication required")
        return
    }
    userID, err := auth.ValidateJWT(tokenString, cfg.JWTSecret)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
        return
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
        UserID:    userID,
    }

    chirp, err := cfg.DB.CreateChirp(r.Context(), params)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Failed to create chirp")
        return
    }

    respondWithJSON(w, http.StatusCreated, ChirpResponse{
        ID:        chirp.ID,
        CreatedAt: chirp.CreatedAt,
        UpdatedAt: chirp.UpdatedAt,
        Body:      chirp.Body,
        UserID:    chirp.UserID,
    })
}

func (cfg *APIConfig) getAllChirpsHandler(w http.ResponseWriter, r *http.Request) {

    var chirps []database.Chirp
    var err error
    
    authorIDStr := r.URL.Query().Get("author_id")

    sortOrder := r.URL.Query().Get("sort")
    if sortOrder == "" || (sortOrder != "asc" && sortOrder != "desc") {
        sortOrder = "asc" 
    }

    if authorIDStr != "" {
        authorID, err := uuid.Parse(authorIDStr)
        if err != nil {
            respondWithError(w, http.StatusBadRequest, "Invalid author ID format")
            return
        }
        
        chirps, err = cfg.DB.GetChirpsByAuthor(r.Context(), authorID)
        if err != nil {
            respondWithError(w, http.StatusInternalServerError, "Failed to retrieve chirps")
            return
        }
    } else {
        chirps, err = cfg.DB.GetAllChirps(r.Context())
        if err != nil {
            respondWithError(w, http.StatusInternalServerError, "Failed to retrieve chirps")
            return
        }
    }

    sort.Slice(chirps, func(i, j int) bool {
        if sortOrder == "asc" {
            return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
        }
        return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
    })

    var response []ChirpResponse
    for _, chirp := range chirps {
        response = append(response, ChirpResponse{
            ID:        chirp.ID,
            CreatedAt: chirp.CreatedAt,
            UpdatedAt: chirp.UpdatedAt,
            Body:      chirp.Body,
            UserID:    chirp.UserID,
        })
    }

    respondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) getChirpHandler(w http.ResponseWriter, r *http.Request) { 
    id := r.PathValue("chirpID")
    chirpID, err := uuid.Parse(id)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
        return
    }

    chirp, err := cfg.DB.GetChirpByID(r.Context(), chirpID)
    if err != nil {
        if err == sql.ErrNoRows {
            respondWithError(w, http.StatusNotFound, "Chirp not found")
            return
        }
        respondWithError(w, http.StatusInternalServerError, "Failed to retrieve chirp")
        return
    }
    response := ChirpResponse{
        ID:        chirp.ID,
        CreatedAt: chirp.CreatedAt,
        UpdatedAt: chirp.UpdatedAt,
        Body:      chirp.Body,
        UserID:    chirp.UserID,
    }
    respondWithJSON(w, http.StatusOK, response)
}

func (cfg *APIConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
    tokenString, err := auth.GetBearerToken(r.Header)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Authentication required")
        return
    }
    
    userID, err := auth.ValidateJWT(tokenString, cfg.JWTSecret)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
        return
    }
    
    chirpIDStr := r.PathValue("chirpID")
    chirpID, err := uuid.Parse(chirpIDStr)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
        return
    }
    
    chirp, err := cfg.DB.GetChirpByID(r.Context(), chirpID)
    if err != nil {
        if err == sql.ErrNoRows {
            respondWithError(w, http.StatusNotFound, "Chirp not found")
            return
        }
        respondWithError(w, http.StatusInternalServerError, "Failed to retrieve chirp")
        return
    }
    
    if chirp.UserID != userID {
        respondWithError(w, http.StatusForbidden, "You can only delete your own chirps")
        return
    }
    
    err = cfg.DB.DeleteChirp(r.Context(), chirpID)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Failed to delete chirp")
        return
    }
    
    w.WriteHeader(http.StatusNoContent)
}