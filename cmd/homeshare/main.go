package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/homeshare/homeshare/internal/config"
	"github.com/homeshare/homeshare/internal/db"
	"github.com/homeshare/homeshare/internal/storage"
)

var version = "0.1.0"

func main() {
	configPath := flag.String("config", "/etc/homeshare/config.yaml", "path to config file")
	showVersion := flag.Bool("version", false, "show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("homeshare v%s\n", version)
		os.Exit(0)
	}

	// Load config
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	// Get secrets
	sessionSecret, err := config.GetSecret(cfg.Security.SessionSecretEnv)
	if err != nil {
		log.Fatalf("Failed to get session secret: %v", err)
	}
	ipHashSalt, err := config.GetSecret(cfg.Security.IPHashSaltEnv)
	if err != nil {
		log.Fatalf("Failed to get IP hash salt: %v", err)
	}

	_ = sessionSecret
	_ = ipHashSalt

	// Open database
	ctx := context.Background()
	database, err := db.Open(ctx, cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Initialize storage
	store, err := storage.NewStorage(cfg.DataDir, cfg.TmpDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	_ = store

	// Create HTTP server
	mux := http.NewServeMux()
	
	// Register handlers (placeholder)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("homeshare v" + version))
	})

	server := &http.Server{
		Addr:         cfg.Listen,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 5 * time.Minute, // Long timeout for uploads
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Starting homeshare v%s on %s", version, cfg.Listen)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-shutdown
	log.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
