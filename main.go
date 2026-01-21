package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/cosmopolitics/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	count   atomic.Int64
	dbquery *database.Queries
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("failed to open db: %s", err)
	}
	dbqueries := database.New(db)
	cfg := &apiConfig{
		dbquery: dbqueries,
	}

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
