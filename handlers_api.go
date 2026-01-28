package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/cosmopolitics/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handler_polka_webhook(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Event string `json:"event"`
		Data  struct {
			User_id string `json:"user_id"`
		} `json:"data"`
	}
	pt, err := auth.GetPolkaToken(r.Header)
	if pt != cfg.polka_key || err != nil {
		respondWithError(w, 
			http.StatusUnauthorized, 
			"bad authorization credentials", 
			err,
		)
		return
	}

	var body request
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&body)
	if err != nil {
		respondWithError(w,
			http.StatusBadRequest,
			"improper body shape",
			err,
		)
		return
	}
	if body.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, "")
		return
	}
	uid, err := uuid.Parse(body.Data.User_id)
	if err != nil {
		respondWithError(w,
			http.StatusBadRequest,
			"improper body shape",
			err,
		)
		return
	}

	_, err = cfg.dbquery.SubcribeUser(r.Context(), uid)
	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusNotFound, "", err)
		return
	} else if err != nil {
		respondWithError(w,
			http.StatusInternalServerError,
			"failed to subscribe user",
			err,
		)
		return
	}
	respondWithJSON(w, http.StatusNoContent, "")
}
