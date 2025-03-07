package main

import (
	"encoding/json"
	"net/http"
	"strings"
    "time"
    "github.com/KrishKoria/Chirpy/internal/auth"
    "github.com/KrishKoria/Chirpy/internal/database"
    "database/sql"
)

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


func (cfg *APIConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
    refreshToken, err := auth.GetBearerToken(r.Header)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Refresh token required")
        return
    }

    tokenData, err := cfg.DB.GetRefreshToken(r.Context(), refreshToken)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Invalid refresh token")
        return
    }

    if time.Now().After(tokenData.ExpiresAt) {
        respondWithError(w, http.StatusUnauthorized, "Refresh token expired")
        return
    }

    if tokenData.RevokedAt.Valid {
        respondWithError(w, http.StatusUnauthorized, "Refresh token revoked")
        return
    }

    newToken, err := auth.MakeJWT(tokenData.UserID, cfg.JWTSecret, time.Hour)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Failed to generate access token")
        return
    }

    type refreshResponse struct {   
        Token string `json:"token"`
    }

    response := refreshResponse{
        Token: newToken,
    }

    respondWithJSON(w, http.StatusOK, response)
}


func (cfg *APIConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {
    refreshToken, err := auth.GetBearerToken(r.Header)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Refresh token required")
        return
    }

    now := time.Now().UTC()

    err = cfg.DB.RevokeRefreshToken(r.Context(), database.RevokeRefreshTokenParams{
        Token:     refreshToken,
        RevokedAt: sql.NullTime{Time: now, Valid: true},
        UpdatedAt: now,
    })

    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Failed to revoke token")
        return
    }

    w.WriteHeader(http.StatusNoContent)
}