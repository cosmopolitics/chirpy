package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (cfg *apiConfig) middlewareMetricsInc(nxt http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.count.Add(1)
		nxt.ServeHTTP(w, req)
	})
}

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
	cfg.count.Swap(0)
}

func handler_chirp_validator(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		Cleaned_body string `json:"cleaned_body"`
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

	params.Body = censor_chirp(params.Body)
	respondWithJSON(w, http.StatusOK, returnVals{
		Cleaned_body: params.Body,
	})
}

func censor_chirp(chirp string) string {
	censored_words := []string{"kerfuffle", "sharbert", "fornax"}
	const censor = "****"
	for _, cword := range censored_words {
		if strings.Contains(strings.ToLower(chirp), cword) {

			split_chirp_at := strings.Index(strings.ToLower(chirp), cword)
			splitchirp := []string{
				chirp[:split_chirp_at],
				chirp[split_chirp_at+len(cword):],
			}
			chirp = strings.Join(splitchirp, censor)
		}
	}
	return chirp
}
