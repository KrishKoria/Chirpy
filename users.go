package main

import (
	"encoding/json"
	"net/http"
    "github.com/KrishKoria/Chirpy/internal/auth"
    "github.com/KrishKoria/Chirpy/internal/database"
	"github.com/lib/pq"
)


func (cfg *APIConfig) UsersHandler(w http.ResponseWriter, r *http.Request) {
    type userRequest struct {
        Email string `json:"email"`
        Password string `json:"password"`
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

    if req.Password == "" {
        respondWithError(w, http.StatusBadRequest, "Password is required")
        return
    }


    hashedPassword, err := auth.HashPassword(req.Password)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Failed to process password")
        return
    }


    user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
        Email:    req.Email,
        HashedPassword: hashedPassword,
    })
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


func (cfg *APIConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
    type loginRequest struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    var req loginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    if req.Email == "" || req.Password == "" {
        respondWithError(w, http.StatusBadRequest, "Email and password are required")
        return
    }

    user, err := cfg.DB.GetUserByEmail(r.Context(), req.Email)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
        return
    }

    err = auth.CheckPasswordHash(req.Password, user.HashedPassword)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
        return
    }

    response := User{
        ID:        user.ID,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
        Email:     user.Email,
    }

    respondWithJSON(w, http.StatusOK, response)
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