package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/mahendrakalkura/saas-blueprint/internal/database"
)

func NewRouter(db *database.DB) *chi.Mux {
	r := chi.NewRouter()

	// Health check endpoints
	healthHandler := NewHealthHandler(db)
	r.Get("/health", healthHandler.Check)
	r.Get("/health/live", healthHandler.Liveness)
	r.Get("/health/ready", healthHandler.Readiness)

	// API v1 routes will be added here
	// Example route groups:
	// r.Route("/auth", func(r chi.Router) {
	//     // Auth routes
	// })
	// r.Route("/users", func(r chi.Router) {
	//     // User routes
	// })
	// r.Route("/organizations", func(r chi.Router) {
	//     // Organization routes
	// })

	return r
}
