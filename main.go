package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	count atomic.Int64
}

func main() {
	const filepathRoot = "."
	const port = "8080"
	cfg := &apiConfig{}

	mux := http.NewServeMux()
	mux.Handle("/app/", 
		cfg.middlewareMetricsInc(
			http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", handler_readiness)
	mux.HandleFunc("GET /admin/metrics", cfg.handler_metrics)
	mux.HandleFunc("POST /admin/reset", cfg.handler_reset)
	mux.HandleFunc("POST /api/validate_chirp", handler_chirp_validator)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

