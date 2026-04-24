package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/abolcerek/HTTPserver/internal/database"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	database *database.Queries
}

type parameters struct {
	Body string `json:"body"`
}
type error_parameters struct {
	Error string `json:"error"`
}
type response_parameters struct {
	Cleaned_Body string `json:"cleaned_body"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1) 
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) HitsHandler(w http.ResponseWriter, r *http.Request) {
	num_hits := cfg.fileserverHits.Load()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	http_response := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", num_hits)
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

func HandlerChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err_params := error_parameters{}
	resp_params := response_parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		err_params.Error = "Something went wrong"
		handleErrors(w, &err_params)
		return
	}
	if len(params.Body) >= 140 {
		err_params.Error = "Chirp is too long"
		handleErrors(w, &err_params)
		return
	}
	resp_params = handleProfanity(params.Body)
	data, err := json.Marshal(resp_params)
	if err != nil {
		log.Printf("Error marshalling JSON")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write(data)
}

func handleProfanity(response string) response_parameters { 
	words := strings.Split(response, " ")
	for i := range words {
		if strings.ToLower(words[i]) == "kerfuffle" || strings.ToLower(words[i]) == "sharbert" || strings.ToLower(words[i]) == "fornax"{
			words[i] = "****"
		}
	}
	cleaned_response := strings.Join(words, " ")
	resp_params := response_parameters{}
	resp_params.Cleaned_Body = cleaned_response
	return resp_params
}

func handleErrors(w http.ResponseWriter, err_params *error_parameters) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(400)
	data, err := json.Marshal(err_params)
	if err != nil {
		log.Printf("Error marshalling JSON")
		return
	}
	w.Write(data) 
}

func main() {
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	mux := http.NewServeMux()
	server := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}
	apiCfg := apiConfig {}
	apiCfg.database = database.New(db)
	fs := http.FileServer(http.Dir("."))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", fs)))
	mux.HandleFunc("GET /admin/metrics", apiCfg.HitsHandler)
	mux.HandleFunc("GET /api/healthz", HealthzHandler)
	mux.HandleFunc("POST /api/validate_chirp", HandlerChirp)
	mux.HandleFunc("POST /admin/reset", apiCfg.HandlerReset)
	server.ListenAndServe()
}