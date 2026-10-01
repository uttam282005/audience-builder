package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mable/audience-builder/backend/internal/api"
	"github.com/mable/audience-builder/backend/internal/evaluator"
	"github.com/mable/audience-builder/backend/internal/store"
)

func main() {
	port := flag.Int("port", 8080, "HTTP server listening port")
	dbPath := flag.String("db", "audience.db", "Path to SQLite database file (or :memory:)")
	autoSeed := flag.Bool("seed", true, "Automatically seed synthetic data if database is empty")
	resetSeed := flag.Bool("reset-seed", false, "Reset existing data and reseed synthetic dataset")
	flag.Parse()

	// Check environment variable overrides
	if envPort := os.Getenv("PORT"); envPort != "" {
		_, _ = fmt.Sscanf(envPort, "%d", port)
	}
	if envDB := os.Getenv("DB_PATH"); envDB != "" {
		*dbPath = envDB
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("starting_backend_server",
		"port", *port,
		"db", *dbPath,
	)

	// 1. Initialize SQLite database
	db, err := store.Open(*dbPath)
	if err != nil {
		logger.Error("failed_to_initialize_database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// 2. Handle seeding
	if *resetSeed {
		logger.Info("resetting_and_reseeding_data")
		if err := db.ResetAndSeed(); err != nil {
			logger.Error("failed_to_reset_and_seed", "error", err)
			os.Exit(1)
		}
	} else if *autoSeed {
		if err := db.SeedData(); err != nil {
			logger.Error("failed_to_seed_database", "error", err)
			os.Exit(1)
		}
		logger.Info("database_seeded_successfully")
	}

	// 3. Initialize Evaluator and HTTP Router
	eval := evaluator.New(db)
	router := api.NewServer(logger, eval)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 4. Start HTTP Server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server_listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// 5. Graceful shutdown on OS signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		logger.Error("server_fatal_error", "error", err)
		os.Exit(1)
	case sig := <-stop:
		logger.Info("shutdown_signal_received", "signal", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("error_during_server_shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("server_shutdown_cleanly")
}
