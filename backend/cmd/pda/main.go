// PDA Service — serves the PDA REST API for warehouse operators.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	port := os.Getenv("PDA_PORT")
	if port == "" {
		port = "8081"
	}

	// TODO: Initialize database connection (PostgreSQL)
	// TODO: Initialize Redis connection
	// TODO: Initialize repositories
	// TODO: Initialize services
	// TODO: Initialize API router (chi/v5) with middleware
	// TODO: Register PDA API routes

	// Placeholder server for now
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","service":"pda","version":"0.1.0"}`)
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down PDA service...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("PDA service forced to shutdown: %v", err)
		}
		log.Println("PDA service stopped")
	}()

	log.Printf("PDA service starting on :%s (http://localhost:%s/health)", port, port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("PDA service failed: %v", err)
	}
}
