package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/mahendrakalkura/saas-blueprint/internal/auth"
	"github.com/mahendrakalkura/saas-blueprint/internal/config"
	"github.com/mahendrakalkura/saas-blueprint/internal/database"
	appMiddleware "github.com/mahendrakalkura/saas-blueprint/internal/middleware"
	"github.com/mahendrakalkura/saas-blueprint/internal/mfa"
	"github.com/mahendrakalkura/saas-blueprint/internal/oauth"
	"github.com/mahendrakalkura/saas-blueprint/internal/payment"
	"github.com/mahendrakalkura/saas-blueprint/internal/repository"
	"github.com/mahendrakalkura/saas-blueprint/internal/storage"
	"github.com/mahendrakalkura/saas-blueprint/internal/websocket"
	"github.com/mahendrakalkura/saas-blueprint/internal/worker"
)

func NewRouter(db *database.DB, workerClient *worker.Client, storageService *storage.Service, paymentService *payment.Service, wsHub *websocket.Hub, cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	// Repositories
	userRepo := repository.NewUserRepository(db.DB)
	sessionRepo := repository.NewSessionRepository(db.DB)
	fileRepo := repository.NewFileRepository(db.DB)
	subRepo := repository.NewSubscriptionRepository(db.DB)
	orgRepo := repository.NewOrganizationRepository(db.DB)
	memberRepo := repository.NewOrganizationMemberRepository(db.DB)
	notificationRepo := repository.NewNotificationRepository(db.DB)
	oauthRepo := repository.NewOAuthProviderRepository(db.DB)

	// Services
	oauthService := oauth.NewService(&cfg.OAuth)
	mfaService := mfa.NewService("SaaS Blueprint")

	// Handlers
	healthHandler := NewHealthHandler(db)
	authHandler := NewAuthHandler(userRepo, sessionRepo, workerClient, cfg, mfaService)
	fileHandler := NewFileHandler(fileRepo, userRepo, storageService)
	billingHandler := NewBillingHandler(paymentService, subRepo, userRepo, cfg)
	orgHandler := NewOrganizationHandler(orgRepo, memberRepo, userRepo, workerClient)
	monitoringHandler := NewMonitoringHandler(cfg)
	wsHandler := websocket.NewHandler(wsHub)
	notificationHandler := NewNotificationHandler(notificationRepo)
	oauthHandler := NewOAuthHandler(oauthService, userRepo, sessionRepo, oauthRepo, cfg)
	mfaHandler := NewMFAHandler(mfaService, userRepo)

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
		r.Post("/mfa/verify", authHandler.VerifyMFA)

		// OAuth routes
		r.Get("/google", oauthHandler.GoogleLogin)
		r.Get("/google/callback", oauthHandler.GoogleCallback)
		r.Get("/github", oauthHandler.GitHubLogin)
		r.Get("/github/callback", oauthHandler.GitHubCallback)
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
			// Public organization routes (auth only)
			r.Post("/", orgHandler.CreateOrganization)
			r.Get("/", orgHandler.ListOrganizations)
			r.Post("/invitations/{token}/accept", orgHandler.AcceptInvitation)

			// Organization-specific routes (require membership)
			r.Route("/{id}", func(r chi.Router) {
				r.Use(appMiddleware.RequireOrganizationMember(memberRepo))

				// Member-level access (read-only)
				r.Get("/", orgHandler.GetOrganization)
				r.Get("/members", orgHandler.ListMembers)

				// Admin-level access (can manage members and settings)
				r.Group(func(r chi.Router) {
					r.Use(appMiddleware.RequireAdmin())
					r.Put("/", orgHandler.UpdateOrganization)
					r.Post("/members/invite", orgHandler.InviteMember)
					r.Put("/members/{memberID}/role", orgHandler.UpdateMemberRole)
					r.Delete("/members/{memberID}", orgHandler.RemoveMember)
				})

				// Owner-only access (can delete organization)
				r.Group(func(r chi.Router) {
					r.Use(appMiddleware.RequireOwner())
					r.Delete("/", orgHandler.DeleteOrganization)
				})
			})
		})

		// Monitoring routes (TODO: add admin-only middleware)
		r.Route("/monitoring", func(r chi.Router) {
			r.Get("/queues", monitoringHandler.GetQueueStats)
			r.Get("/servers", monitoringHandler.GetServerInfo)
			r.Get("/scheduled", monitoringHandler.GetScheduledTasks)
		})

		// Notification routes
		r.Route("/notifications", func(r chi.Router) {
			r.Get("/", notificationHandler.ListNotifications)
			r.Get("/unread-count", notificationHandler.GetUnreadCount)
			r.Post("/mark-all-read", notificationHandler.MarkAllAsRead)
			r.Post("/{id}/read", notificationHandler.MarkAsRead)
			r.Delete("/{id}", notificationHandler.DeleteNotification)
		})

		// MFA routes
		r.Route("/mfa", func(r chi.Router) {
			r.Get("/status", mfaHandler.GetMFAStatus)
			r.Post("/enable", mfaHandler.EnableMFA)
			r.Post("/verify", mfaHandler.VerifyAndActivateMFA)
			r.Post("/disable", mfaHandler.DisableMFA)
			r.Post("/backup-codes/regenerate", mfaHandler.RegenerateBackupCodes)
		})

		// WebSocket routes
		r.Get("/ws", wsHandler.ServeWS)
		r.Get("/ws/stats", wsHandler.GetStats)
	})

	return r
}
