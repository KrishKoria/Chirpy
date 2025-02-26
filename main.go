package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)


type apiConfig struct {
	fileserverHits atomic.Int32
}


func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cfg.fileserverHits.Add(1)
        next.ServeHTTP(w, r)
    })
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8");
	w.WriteHeader(http.StatusOK);
	w.Write([]byte("OK"));
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Hits: %d", cfg.fileserverHits.Load())
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
    cfg.fileserverHits.Store(0)
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Hits reset to 0"))
}

func main()  {
	mux := http.NewServeMux();
	cfg := &apiConfig{}
	mux.Handle("/app/", http.StripPrefix("/app", cfg.middlewareMetricsInc(http.FileServer(http.Dir("./app")))));
	mux.HandleFunc("GET /healthz", readinessHandler);
	mux.HandleFunc("GET /metrics", cfg.metricsHandler);
	mux.HandleFunc("POST /reset", cfg.resetHandler);
	server:= &http.Server{
		Addr: ":8080",
		Handler: mux,
	}
	server.ListenAndServe();
}