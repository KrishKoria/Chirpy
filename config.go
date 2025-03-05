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
}

type User struct {
    ID        uuid.UUID `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Email     string    `json:"email"`
}