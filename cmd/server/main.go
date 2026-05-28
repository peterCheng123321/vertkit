package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vertkit/vertkit/internal/api"
	"github.com/vertkit/vertkit/internal/storage"
	"github.com/vertkit/vertkit/internal/storage/memory"
	"github.com/vertkit/vertkit/internal/storage/postgres"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("VertKit starting (global CRM/ERP framework)")

	stores, closeStores := mustOpenStores()
	defer closeStores()

	serviceToken := os.Getenv("VERTKIT_SERVICE_TOKEN")
	if serviceToken == "" {
		log.Println("VERTKIT_SERVICE_TOKEN is not set; service-token auth is disabled")
	}
	srv := api.NewServer(stores, api.WithServiceToken(serviceToken))

	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		fmt.Sscanf(p, "%d", &port)
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: srv.Router(),
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("HTTP server listening on :%d", port)
		log.Printf("Try: curl -X POST http://localhost:%d/tenants -d '{\"id\":\"t1\",\"name\":\"Acme Corp\",\"default_currency\":\"USD\"}'", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	_ = srv.Shutdown(shutdownCtx)

	log.Println("VertKit stopped")
}

func mustOpenStores() (*storage.Stores, func()) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Println("Using in-memory storage; set DATABASE_URL for Postgres")
		return memory.NewStores(), func() {}
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("open postgres: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		log.Fatalf("ping postgres: %v", err)
	}
	if err := postgres.ApplySchema(ctx, db); err != nil {
		_ = db.Close()
		log.Fatalf("apply postgres schema: %v", err)
	}
	log.Println("Using Postgres storage")
	return postgres.NewStores(db), func() {
		_ = db.Close()
	}
}
