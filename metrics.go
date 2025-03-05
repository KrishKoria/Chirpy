package main

import (
	"fmt"
	"net/http"
)
func (cfg *APIConfig) middlewareMetricsInc(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cfg.FileserverHits.Add(1)
        next.ServeHTTP(w, r)
    })
}

func ReadinessHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func (cfg *APIConfig) MetricsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, `
        <html>
          <body>
            <h1>Welcome, Chirpy Admin</h1>
            <p>Chirpy has been visited %d times!</p>
          </body>
        </html>`, cfg.FileserverHits.Load())
}