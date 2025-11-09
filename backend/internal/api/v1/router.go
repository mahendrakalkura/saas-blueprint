package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/mahendrakalkura/saas-blueprint/internal/auth"
	"github.com/mahendrakalkura/saas-blueprint/internal/config"
	"github.com/mahendrakalkura/saas-blueprint/internal/database"
	"github.com/mahendrakalkura/saas-blueprint/internal/payment"
	"github.com/mahendrakalkura/saas-blueprint/internal/repository"
	"github.com/mahendrakalkura/saas-blueprint/internal/storage"
	"github.com/mahendrakalkura/saas-blueprint/internal/worker"
)

func NewRouter(db *database.DB, workerClient *worker.Client, storageService *storage.Service, paymentService *payment.Service, cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	// Repositories
	userRepo := repository.NewUserRepository(db.DB)
	sessionRepo := repository.NewSessionRepository(db.DB)
	fileRepo := repository.NewFileRepository(db.DB)
	subRepo := repository.NewSubscriptionRepository(db.DB)
	orgRepo := repository.NewOrganizationRepository(db.DB)
	memberRepo := repository.NewOrganizationMemberRepository(db.DB)

	// Handlers
	healthHandler := NewHealthHandler(db)
	authHandler := NewAuthHandler(userRepo, sessionRepo, workerClient, cfg)
	fileHandler := NewFileHandler(fileRepo, userRepo, storageService)
	billingHandler := NewBillingHandler(paymentService, subRepo, userRepo, cfg)
	orgHandler := NewOrganizationHandler(orgRepo, memberRepo, userRepo, workerClient)
	monitoringHandler := NewMonitoringHandler(cfg)

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

	// Stripe webhook (public, but verified)
	r.Post("/webhooks/stripe", billingHandler.HandleWebhook)

	// Protected routes (require authentication)
	r.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware(&cfg.JWT))

		r.Get("/auth/me", authHandler.Me)

		// File routes
		r.Route("/files", func(r chi.Router) {
			r.Post("/", fileHandler.UploadFile)
			r.Post("/avatar", fileHandler.UploadAvatar)
			r.Get("/", fileHandler.ListFiles)
			r.Get("/{id}", fileHandler.GetFile)
			r.Delete("/{id}", fileHandler.DeleteFile)
		})

		// Billing routes
		r.Route("/billing", func(r chi.Router) {
			r.Post("/checkout", billingHandler.CreateCheckoutSession)
			r.Get("/subscription", billingHandler.GetSubscription)
			r.Post("/subscription/cancel", billingHandler.CancelSubscription)
			r.Post("/subscription/reactivate", billingHandler.ReactivateSubscription)
			r.Post("/subscription/update", billingHandler.UpdateSubscription)
			r.Post("/portal", billingHandler.CreatePortalSession)
		})

		// Organization routes
		r.Route("/organizations", func(r chi.Router) {
			r.Post("/", orgHandler.CreateOrganization)
			r.Get("/", orgHandler.ListOrganizations)
			r.Get("/{id}", orgHandler.GetOrganization)
			r.Put("/{id}", orgHandler.UpdateOrganization)
			r.Delete("/{id}", orgHandler.DeleteOrganization)

			// Member management
			r.Get("/{id}/members", orgHandler.ListMembers)
			r.Post("/{id}/members/invite", orgHandler.InviteMember)
			r.Post("/invitations/{token}/accept", orgHandler.AcceptInvitation)
			r.Put("/{id}/members/{memberID}/role", orgHandler.UpdateMemberRole)
			r.Delete("/{id}/members/{memberID}", orgHandler.RemoveMember)
		})

		// Monitoring routes (TODO: add admin-only middleware)
		r.Route("/monitoring", func(r chi.Router) {
			r.Get("/queues", monitoringHandler.GetQueueStats)
			r.Get("/servers", monitoringHandler.GetServerInfo)
			r.Get("/scheduled", monitoringHandler.GetScheduledTasks)
		})
	})

	return r
}
