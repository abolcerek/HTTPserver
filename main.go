package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/abolcerek/HTTPserver/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	database *database.Queries
	platform string
}
type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}
type parameters struct {
	Body string `json:"body"`
	ID uuid.UUID `json:"user_id"`
}
type error_parameters struct {
	Error string `json:"error"`
}
type response_parameters struct {
	Cleaned_Body string `json:"cleaned_body"`
}

type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body     string    `json:"body"`
		UserID     uuid.UUID    `json:"user_id"`
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
	if cfg.platform != "dev" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(403)
		http_response := "403 Forbidden"
		w.Write([]byte(http_response))
		return
	}
	cfg.fileserverHits.Store(0)
	ctx := context.Background()
	err := cfg.database.ResetUsers(ctx)
	if err != nil {
		err_params := error_parameters{}
		err_params.Error = "Something went wrong"
		handleErrors(w, &err_params)
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	http_response := "200 OK"
	w.Write([]byte(http_response))
}
func (cfg *apiConfig) CreateUser(w http.ResponseWriter, r *http.Request){
	type request_params struct {
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	req_params := request_params{}
	err_params := error_parameters{}
	err := decoder.Decode(&req_params)
	user_params := database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email: req_params.Email,
	}
	if err != nil {
		err_params.Error = "Something went wrong"
		handleErrors(w, &err_params)
		return
	}
	database_user, err := cfg.database.CreateUser(r.Context(), user_params)
	if err != nil {
		err_params.Error = "Something went wrong"
		handleErrors(w, &err_params)
		return
	}
	user := User{
		ID: database_user.ID,
		CreatedAt: database_user.CreatedAt,
		UpdatedAt: database_user.UpdatedAt,
		Email: database_user.Email,
	}
	data, err := json.Marshal(user)
	if err != nil {
		err_params.Error = "Something went wrong"
		handleErrors(w, &err_params)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(201)
	w.Write(data)
}

func (cfg *apiConfig) HandlerChirp(w http.ResponseWriter, r *http.Request) {
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
	ctx := context.Background()
	chirp_params := database.CreateChirpParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body: resp_params.Cleaned_Body,
		UserID: params.ID,
	}
	chirp, err := cfg.database.CreateChirp(ctx, chirp_params)
	if err != nil {
		err_params.Error = "Something went wrong"
		handleErrors(w, &err_params)
		return
	}
	chirp_response := Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	}
	data, err := json.Marshal(chirp_response)
	if err != nil {
		log.Printf("Error marshalling JSON")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(201)
	w.Write(data)
}

func (cfg *apiConfig) HandlerGetChirps(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	chirps, err := cfg.database.GetChirps(ctx)
	err_params := error_parameters{}
	if err != nil {
		err_params.Error = "Something went wrong"
		handleErrors(w, &err_params)
		return
	}
	chirps_response := []Chirp{}
	for i := range chirps {
		chirp_response := Chirp{
			ID: chirps[i].ID,
			CreatedAt: chirps[i].CreatedAt,
			UpdatedAt: chirps[i].UpdatedAt,
			Body: chirps[i].Body,
			UserID: chirps[i].UserID,
		}
		chirps_response = append(chirps_response, chirp_response)
	}
	data, err := json.Marshal(chirps_response)
	if err != nil {
		log.Printf("Error marshalling JSON")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *apiConfig) HandlerGetChirp(w http.ResponseWriter, r *http.Request) {
	err_params := error_parameters{}
	chirp_id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		err_params.Error = "Something went wrong"
		handleErrors(w, &err_params)
		return
	}
	ctx := context.Background()
	chirp, err := cfg.database.GetChirp(ctx, chirp_id)
	if err != nil {
		err_params.Error = "No Chirp Found"
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(404)
		data, err := json.Marshal(err_params)
		if err != nil {
			log.Printf("Error marshalling JSON")
			return
		}
		w.Write(data) 
		return
	}
	chirp_response := Chirp{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
	}
	data, err := json.Marshal(chirp_response)
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
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
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
	apiCfg.platform = platform
	fs := http.FileServer(http.Dir("."))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", fs)))
	mux.HandleFunc("GET /admin/metrics", apiCfg.HitsHandler)
	mux.HandleFunc("POST /api/users", apiCfg.CreateUser)
	mux.HandleFunc("GET /api/healthz", HealthzHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.HandlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.HandlerGetChirp)
	mux.HandleFunc("POST /api/chirps", apiCfg.HandlerChirp)
	mux.HandleFunc("POST /admin/reset", apiCfg.HandlerReset)
	server.ListenAndServe()
}