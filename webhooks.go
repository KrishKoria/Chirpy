package main

import (
    "encoding/json"
    "net/http"

    "github.com/google/uuid"
    "github.com/KrishKoria/Chirpy/internal/auth"
)

func (cfg *APIConfig) polkaWebhookHandler(w http.ResponseWriter, r *http.Request) {
    apiKey, err := auth.GetAPIKey(r.Header)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "API key required")
        return
    }

    if apiKey != cfg.PolkaKey {
        respondWithError(w, http.StatusUnauthorized, "Invalid API key")
        return
    }

    type webhookData struct {
        UserID string `json:"user_id"`
    }

    type webhookPayload struct {
        Event string      `json:"event"`
        Data  webhookData `json:"data"`
    }

    var payload webhookPayload
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    if payload.Event != "user.upgraded" {
        w.WriteHeader(http.StatusNoContent)
        return
    }

    userID, err := uuid.Parse(payload.Data.UserID)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid user ID format")
        return
    }

    _, err = cfg.DB.UpgradeUserToChirpyRed(r.Context(), userID)
    if err != nil {
        respondWithError(w, http.StatusNotFound, "User not found")
        return
    }

    w.WriteHeader(http.StatusNoContent)
}