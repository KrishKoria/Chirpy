package main

import (
	"encoding/json"
	"net/http"
	"strings"
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
