package main

import (
	"sync/atomic"
	"time"

	"github.com/KrishKoria/Chirpy/internal/database"
	"github.com/google/uuid"
)

type APIConfig struct {
    FileserverHits atomic.Int32
    DB             *database.Queries
    Platform       string
    JWTSecret      string
}

type User struct {
    ID        uuid.UUID `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Email     string    `json:"email"`
}

type ChirpResponse struct {
    ID        uuid.UUID `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Body      string    `json:"body"`
    UserID    uuid.UUID `json:"user_id"`
}