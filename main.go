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

func (cfg *apiConfig) HitsHandler(w http.ResponseWriter, r *http.Request) {
	num_hits := cfg.fileserverHits.Load()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	http_response := fmt.Sprintf("Hits: %v", num_hits)
	w.Write([]byte(http_response))
}

func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	http_response := "200 OK"
	w.Write([]byte(http_response))
}

func (cfg *apiConfig) HandlerReset(w http.ResponseWriter, r *http.Request){
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	http_response := "200 OK"
	w.Write([]byte(http_response))
}

func main() {
	mux := http.NewServeMux()
	server := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}
	apiCfg := apiConfig {}
	fs := http.FileServer(http.Dir("."))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", fs)))
	mux.HandleFunc("GET /api/metrics", apiCfg.HitsHandler)
	mux.HandleFunc("GET /api/healthz", HealthzHandler)
	mux.HandleFunc("POST /api/reset", apiCfg.HandlerReset)
	server.ListenAndServe()
}