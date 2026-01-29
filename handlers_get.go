package main

import (
	"fmt"
	"net/http"
	"slices"

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

func (cfg *apiConfig) handler_get_chirps(w http.ResponseWriter, req *http.Request) {
	primary_id := req.URL.Query().Get("author_id")
	sort_dir := req.URL.Query().Get("sort")

	var chirps []database.Chirp
	var err error
	if primary_id == "" {
		chirps, err = cfg.dbquery.GetAllChirps(req.Context())
		if err != nil {
			respondWithError(w,
				http.StatusInternalServerError,
				"db failed",
				err,
			)
			return
		}
	} else {
		uid, err := uuid.Parse(primary_id)
		if err != nil {
			respondWithError(w,
				http.StatusBadRequest,
				"malformed query",
				err,
			)
			return
		}
		chirps, err = cfg.dbquery.GetUsersChirps(req.Context(), uid)
		if err != nil {
			respondWithError(w,
				http.StatusInternalServerError,
				"db  failed",
				err,
			)
			return
		}
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

	if sort_dir == "asc" {
		slices.SortFunc(chirpsjtags, func(a, b Chirp) int {
			return a.CreatedAt.Compare(b.CreatedAt)
		})
	} else if sort_dir == "desc" {
		slices.SortFunc(chirpsjtags, func(a, b Chirp) int {
			return b.CreatedAt.Compare(a.CreatedAt)
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
