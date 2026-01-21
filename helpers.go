package main

import (
	"net/http"
	"strings"
)

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

func (cfg *apiConfig) middlewareMetricsInc(nxt http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.count.Add(1)
		nxt.ServeHTTP(w, req)
	})
}
