package main

import (
	"encoding/json"
	"net/http"
	"time"
    "database/sql"
	"github.com/KrishKoria/Chirpy/internal/database"
	"github.com/google/uuid"
    "github.com/KrishKoria/Chirpy/internal/auth"

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

    chirps, err := cfg.DB.GetAllChirps(r.Context())
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Failed to retrieve chirps")
        return
    }

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