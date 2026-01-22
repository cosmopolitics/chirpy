package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/cosmopolitics/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	count   atomic.Int64
	dbquery *database.Queries
}

type Chirp struct {
	ID        uuid.UUID `json:"chirp_id"`
	Uid       uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
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
	mux.HandleFunc("POST /api/chirps", cfg.handler_add_chirp)
	mux.HandleFunc("GET /api/chirps", cfg.handler_get_chirps)
	mux.HandleFunc("POST /api/users", cfg.handler_create_user)
	mux.HandleFunc("GET /api/chirps/{chirp_id}", cfg.handler_get_a_chirp)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
