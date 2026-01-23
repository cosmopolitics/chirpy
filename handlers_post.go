package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/cosmopolitics/chirpy/internal/auth"
	"github.com/cosmopolitics/chirpy/internal/database"
	"github.com/google/uuid"
)

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
		Body string `json:"body"`
	}

	//
	bt, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "no authorization provided", err)
		return
	}
	uid, err := auth.ValidateJwt(bt, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "bad authorization provided", fmt.Errorf("%s: key: %s", err, bt))
		return
	}

	//
	var params parameters
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't decode chirp", err)
		return
	}

	const maxchirplen = 140
	if len(params.Body) > maxchirplen {
		respondWithError(w, http.StatusBadRequest, "max chirp length exceeded", nil)
		return
	}

	verified_chirp := censor_chirp(params.Body)
	chirp, err := cfg.dbquery.AddChirp(req.Context(), database.AddChirpParams{
		Body: verified_chirp,
		UserID:  uid,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to decode db response", err)
		return
	}
	//

	chirped := Chirp{
		chirp.ID,
		chirp.UserID,
		chirp.CreatedAt,
		chirp.UpdatedAt,
		chirp.Body,
	}
	respondWithJSON(w, http.StatusCreated, chirped)
}

func (cfg *apiConfig) handler_login(w http.ResponseWriter, r *http.Request) {
	type login struct {
		Email              string `json:"email"`
		Password           string `json:"password"`
		Expires_in_seconds int    `json:"expires_in_seconds"`
	}
	type response struct {
		ID         uuid.UUID `json:"id"`
		Created_at time.Time `json:"created_at"`
		Updated_at time.Time `json:"updated_at"`
		Email      string    `json:"email"`
		Token      string    `json:"token"`
	}

	decoder := json.NewDecoder(r.Body)
	var parsed_login login
	err := decoder.Decode(&parsed_login)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to decode body", err)
		return
	}
	if parsed_login.Expires_in_seconds == 0 {
		parsed_login.Expires_in_seconds = 60 * 60
	}

	//
	user, err := cfg.dbquery.GetUserByEmail(r.Context(), parsed_login.Email)
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusBadRequest, "no user with that email", err)
		return

	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error fetching user", err)
		return
	}

	//
	correct_password, err := auth.CheckPassword(parsed_login.Password, user.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error processing password", err)
		return
	}

	if !correct_password {
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password", nil)
		return
	}

	jtoken, err := auth.MakeJwt(
		user.ID,
		cfg.secret,
		time.Second * time.Duration(parsed_login.Expires_in_seconds),
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to make jwt ;-;", err)
		return
	}

	//
	respondWithJSON(w, http.StatusOK, response{
		ID:         user.ID,
		Created_at: user.CreatedAt,
		Updated_at: user.UpdatedAt,
		Email:      user.Email,
		Token:      jtoken,
	})
}
