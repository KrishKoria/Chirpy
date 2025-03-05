package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/KrishKoria/Chirpy/internal/database"
	"github.com/google/uuid"
)


func (cfg *APIConfig) chirpsHandler(w http.ResponseWriter, r *http.Request) {
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

    chirp, err := cfg.DB.CreateChirp(r.Context(), params)
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

func (cfg *APIConfig) getAllChirpsHandler(w http.ResponseWriter, r *http.Request) {
    type chirpResponse struct {
        ID        uuid.UUID `json:"id"`
        CreatedAt time.Time `json:"created_at"`
        UpdatedAt time.Time `json:"updated_at"`
        Body      string    `json:"body"`
        UserID    uuid.UUID `json:"user_id"`
    }

    chirps, err := cfg.DB.GetAllChirps(r.Context())
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Failed to retrieve chirps")
        return
    }

    var response []chirpResponse
    for _, chirp := range chirps {
        response = append(response, chirpResponse{
            ID:        chirp.ID,
            CreatedAt: chirp.CreatedAt,
            UpdatedAt: chirp.UpdatedAt,
            Body:      chirp.Body,
            UserID:    chirp.UserID,
        })
    }

    respondWithJSON(w, http.StatusOK, response)
}