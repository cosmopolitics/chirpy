package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/cosmopolitics/chirpy/internal/auth"
	"github.com/cosmopolitics/chirpy/internal/database"
	"github.com/google/uuid"
)

func handler_readiness(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) handler_metrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(
		fmt.Sprintf(`
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.count.Load())))
}

func (cfg *apiConfig) handler_reset(w http.ResponseWriter, req *http.Request) {
	dev_status := os.Getenv("PLATFORM")
	if dev_status != "dev" {
		respondWithError(w, http.StatusForbidden, "", nil)
		return
	}
	cfg.count.Swap(0)
	cfg.dbquery.Reset(req.Context())

}

func (cfg *apiConfig) handler_create_user(w http.ResponseWriter, req *http.Request) {
	type register_user struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type User struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	var jsonBody register_user
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&jsonBody)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to read request body", err)
		return
	}

	hash, err := auth.HashPassword(jsonBody.Password)
	user, err := cfg.dbquery.AddUser(req.Context(),
		database.AddUserParams{
			Email:    jsonBody.Email,
			Password: hash,
		})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to add user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}

func (cfg *apiConfig) handler_add_chirp(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		User_id string `json:"user_id"`
		Body    string `json:"body"`
	}

	var params parameters
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't decode chirp", err)
		return
	}

	const maxchirplen = 140
	if len(params.Body) > maxchirplen {
		respondWithError(w, http.StatusBadRequest, "max chirp length exceeded", nil)
		return
	}
	uid, err := uuid.Parse(params.User_id)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "malformed user_id parameter", err)
		return
	}
	verified_chirp := censor_chirp(params.Body)
	chirp, err := cfg.dbquery.AddChirp(req.Context(), database.AddChirpParams{
		Body: verified_chirp,
		Uid:  uid,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to decode db response", err)
		return
	}

	chirped := Chirp{
		chirp.ID,
		chirp.Uid,
		chirp.CreatedAt,
		chirp.UpdatedAt,
		chirp.Body,
	}
	respondWithJSON(w, http.StatusCreated, chirped)
}

func (cfg *apiConfig) handler_get_chirps(w http.ResponseWriter, req *http.Request) {
	chirps, err := cfg.dbquery.GetAllChirps(req.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "db failed", err)
		return
	}

	var chirpsjtags []Chirp
	for _, c := range chirps {
		chirpsjtags = append(chirpsjtags, Chirp{
			c.ID,
			c.Uid,
			c.CreatedAt,
			c.UpdatedAt,
			c.Body,
		})
	}

	respondWithJSON(w, http.StatusOK, chirpsjtags)
}

func (cfg *apiConfig) handler_get_a_chirp(w http.ResponseWriter, req *http.Request) {
	chirp_id := req.PathValue("chirp_id")
	cid, err := uuid.Parse(chirp_id)
	if err != nil {
		respondWithError(w,
			http.StatusBadRequest,
			fmt.Sprintf("malformed chirp id: %s", chirp_id),
			err,
		)
		return
	}
	chirp, err := cfg.dbquery.GetChirpById(req.Context(), cid)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		chirp.ID,
		chirp.Uid,
		chirp.CreatedAt,
		chirp.UpdatedAt,
		chirp.Body,
	})
}
