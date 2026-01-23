package main

import (
	"fmt"
	"net/http"

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
			c.UserID,
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
		chirp.UserID,
		chirp.CreatedAt,
		chirp.UpdatedAt,
		chirp.Body,
	})
}
