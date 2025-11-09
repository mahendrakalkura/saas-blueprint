package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/mahendrakalkura/saas-blueprint/internal/auth"
	"github.com/mahendrakalkura/saas-blueprint/internal/config"
	"github.com/mahendrakalkura/saas-blueprint/internal/database"
	"github.com/mahendrakalkura/saas-blueprint/internal/repository"
)

func NewRouter(db *database.DB, cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	// Repositories
	userRepo := repository.NewUserRepository(db.DB)
	sessionRepo := repository.NewSessionRepository(db.DB)

	// Handlers
	healthHandler := NewHealthHandler(db)
	authHandler := NewAuthHandler(userRepo, sessionRepo, cfg)

	// Health check endpoints
	r.Get("/health", healthHandler.Check)
	r.Get("/health/live", healthHandler.Liveness)
	r.Get("/health/ready", healthHandler.Readiness)

	// Auth routes (public)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/logout", authHandler.Logout)
		r.Post("/verify-email", authHandler.VerifyEmail)
		r.Post("/request-password-reset", authHandler.RequestPasswordReset)
		r.Post("/reset-password", authHandler.ResetPassword)
	})

	// Protected routes (require authentication)
	r.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware(&cfg.JWT))

		r.Get("/auth/me", authHandler.Me)

		// Additional protected routes will be added here
		// r.Route("/users", func(r chi.Router) {
		//     // User management routes
		// })
		// r.Route("/organizations", func(r chi.Router) {
		//     // Organization routes
		// })
	})

	return r
}
