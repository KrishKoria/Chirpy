package main

import (
	"encoding/json"
	"net/http"

	"github.com/lib/pq"
)


func (cfg *APIConfig) UsersHandler(w http.ResponseWriter, r *http.Request) {
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

    user, err := cfg.DB.CreateUser(r.Context(), req.Email)
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



func (cfg *APIConfig) ResetHandler(w http.ResponseWriter, r *http.Request) {
    if cfg.Platform != "dev" {
        w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        w.WriteHeader(http.StatusForbidden)
        w.Write([]byte("Forbidden"))
        return
    }

    err := cfg.DB.DeleteAllUsers(r.Context())
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Failed to delete users")
        return
    }

    cfg.FileserverHits.Store(0)
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("All users deleted and hits reset to 0"))
}