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
	count     atomic.Int64
	dbquery   *database.Queries
	secret    string
	polka_key string
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	Uid       uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	err := godotenv.Load()
	if err != nil {
		type inter interface{}
		nilinter := new(inter)
		panic(nilinter)
	}
	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("failed to open db: %s", err)
	}
	dbqueries := database.New(db)
	cfg := &apiConfig{
		dbquery:   dbqueries,
		secret:    os.Getenv("JWT_SIGNING_KEY"),
		polka_key: os.Getenv("POLKA_KEY"),
	}

	mux := http.NewServeMux()
	mux.Handle("GET /app/",
		cfg.middlewareMetricsInc(
			http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))

	mux.HandleFunc("GET /api/chirps", cfg.handler_get_chirps)
	mux.HandleFunc("GET /api/healthz", handler_readiness)
	mux.HandleFunc("GET /admin/metrics", cfg.handler_metrics)
	mux.HandleFunc("GET /api/chirps/{chirp_id}", cfg.handler_get_a_chirp)

	mux.HandleFunc("POST /api/users", cfg.handler_create_user)
	mux.HandleFunc("POST /api/login", cfg.handler_login)
	mux.HandleFunc("POST /api/chirps", cfg.handler_add_chirp)
	mux.HandleFunc("POST /admin/reset", cfg.handler_reset)
	mux.HandleFunc("POST /api/refresh", cfg.handler_refresh)
	mux.HandleFunc("POST /api/revoke", cfg.handler_revoke)
	mux.HandleFunc("DELETE /api/chirps/{chirp_id}", cfg.handler_delete_chirp)

	mux.HandleFunc("POST /api/polka/webhooks", cfg.handler_polka_webhook)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: time.Duration(time.Second * 10),
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
