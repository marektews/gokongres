package main

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gokongres/api"
	"gokongres/db"
	"gokongres/sessions"
)

func main() {
	host := flag.String("host", "127.0.0.1", "host to bind the server to")
	port := flag.Int("port", 1977, "port to bind the server to")
	flag.Parse()

	// Initialize session store
	// authKey := []byte("auth-key-change-me-in-production")
	// encKey := []byte("enc-key-change-me-in-production-1234567890ab")
	authKey := sha256.Sum256([]byte("development-auth-key"))
	encKey := sha256.Sum256([]byte("development-enc-key"))
	if err := sessions.InitSessions(authKey, encKey); err != nil {
		log.Fatalf("Failed to initialize sessions: %v", err)
	}

	log.Printf("Connecting to MongoDB...")
	err := db.Connect(context.Background(), "")
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		log.Printf("Disconnecting from MongoDB...")
		if err := db.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from database: %v", err)
		}
	}()

	api.RegisterHandlers(*host, *port)

	addr := fmt.Sprintf("%s:%d", *host, *port)
	srv := &http.Server{
		Addr:    addr,
		Handler: nil,
	}

	go func() {
		log.Printf("Starting HTTP server on http://%s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received, shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exiting")
}
