package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/mahendrakalkura/saas-blueprint/internal/api/v1"
	"github.com/mahendrakalkura/saas-blueprint/internal/config"
	"github.com/mahendrakalkura/saas-blueprint/internal/database"
	"github.com/mahendrakalkura/saas-blueprint/internal/email"
	"github.com/mahendrakalkura/saas-blueprint/internal/logger"
	appMiddleware "github.com/mahendrakalkura/saas-blueprint/internal/middleware"
	"github.com/mahendrakalkura/saas-blueprint/internal/repository"
	"github.com/mahendrakalkura/saas-blueprint/internal/storage"
	"github.com/mahendrakalkura/saas-blueprint/internal/worker"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	log := logger.New(cfg.Server.Environment)
	log.Info().Str("environment", cfg.Server.Environment).Msg("Starting application")

	// Initialize database
	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	log.Info().Msg("Database connection established")

	// Initialize repositories
	sessionRepo := repository.NewSessionRepository(db.DB)

	// Initialize storage service
	storageService, err := storage.NewService(&cfg.Storage)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize storage service")
	}
	log.Info().Msg("Storage service initialized")

	// Initialize email service
	emailService := email.NewService(&cfg.Email)
	log.Info().Msg("Email service initialized")

	// Initialize worker client
	workerClient := worker.NewClient(cfg.Redis.Addr)
	log.Info().Msg("Worker client initialized")

	// Initialize worker server for background job processing
	workerServer := worker.NewServer(cfg.Redis.Addr, emailService, sessionRepo, &log.Logger)

	// Start worker server in a goroutine
	go func() {
		log.Info().Msg("Worker server starting")
		if err := workerServer.Start(); err != nil {
			log.Error().Err(err).Msg("Worker server failed")
		}
	}()
	defer workerServer.Shutdown()

	// Initialize and start scheduler for periodic tasks
	scheduler := worker.NewScheduler(cfg.Redis.Addr, &log.Logger)
	if err := scheduler.RegisterTasks(); err != nil {
		log.Fatal().Err(err).Msg("Failed to register scheduled tasks")
	}

	go func() {
		log.Info().Msg("Scheduler starting")
		if err := scheduler.Start(); err != nil {
			log.Error().Err(err).Msg("Scheduler failed")
		}
	}()
	defer scheduler.Shutdown()

	// Initialize router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(appMiddleware.Logger(log.Logger))
	r.Use(middleware.Recoverer)
	r.Use(appMiddleware.SecurityHeaders)
	r.Use(middleware.Compress(5))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.Server.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// API v1 routes
	r.Mount("/api/v1", v1.NewRouter(db, workerClient, storageService, cfg))

	// Legacy health endpoint for backward compatibility
	healthHandler := v1.NewHealthHandler(db)
	r.Get("/api/health", healthHandler.Check)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Info().Str("port", cfg.Server.Port).Msg("Server starting")
		serverErrors <- srv.ListenAndServe()
	}()

	// Wait for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatal().Err(err).Msg("Server failed to start")

	case sig := <-shutdown:
		log.Info().Str("signal", sig.String()).Msg("Shutdown signal received")

		// Create context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		// Shutdown server gracefully
		if err := srv.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("Graceful shutdown failed")
			if err := srv.Close(); err != nil {
				log.Fatal().Err(err).Msg("Failed to close server")
			}
		}

		log.Info().Msg("Server stopped gracefully")
	}
}
