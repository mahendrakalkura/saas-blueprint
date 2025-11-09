package v1

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mahendrakalkura/saas-blueprint/internal/auth"
	"github.com/mahendrakalkura/saas-blueprint/internal/config"
	"github.com/mahendrakalkura/saas-blueprint/internal/models"
	"github.com/mahendrakalkura/saas-blueprint/internal/payment"
	"github.com/mahendrakalkura/saas-blueprint/internal/repository"
	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

type BillingHandler struct {
	paymentService *payment.Service
	subRepo        *repository.SubscriptionRepository
	userRepo       *repository.UserRepository
	cfg            *config.Config
}

func NewBillingHandler(paymentService *payment.Service, subRepo *repository.SubscriptionRepository, userRepo *repository.UserRepository, cfg *config.Config) *BillingHandler {
	return &BillingHandler{
		paymentService: paymentService,
		subRepo:        subRepo,
		userRepo:       userRepo,
		cfg:            cfg,
	}
}

type CreateCheckoutSessionRequest struct {
	PriceID   string `json:"price_id" validate:"required"`
	TrialDays int64  `json:"trial_days"`
}

type CheckoutSessionResponse struct {
	URL string `json:"url"`
}

// CreateCheckoutSession creates a Stripe Checkout session
func (h *BillingHandler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateCheckoutSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), userCtx.UserID)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	// Create checkout session
	url, err := h.paymentService.CreateCheckoutSession(payment.CreateCheckoutSessionParams{
		CustomerEmail: user.Email,
		PriceID:       req.PriceID,
		SuccessURL:    h.cfg.Server.FrontendURL + "/billing/success",
		CancelURL:     h.cfg.Server.FrontendURL + "/billing",
		TrialDays:     req.TrialDays,
		Metadata: map[string]string{
			"user_id": userCtx.UserID,
		},
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to create checkout session")
		respondError(w, http.StatusInternalServerError, "Failed to create checkout session")
		return
	}

	respondJSON(w, http.StatusOK, CheckoutSessionResponse{URL: url})
}

// GetSubscription returns the current user's subscription
func (h *BillingHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sub, err := h.subRepo.GetByUserID(r.Context(), userCtx.UserID)
	if err != nil {
		respondError(w, http.StatusNotFound, "No active subscription found")
		return
	}

	respondJSON(w, http.StatusOK, sub)
}

type CancelSubscriptionRequest struct {
	Immediately bool `json:"immediately"`
}

// CancelSubscription cancels the current user's subscription
func (h *BillingHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CancelSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Immediately = false
	}

	sub, err := h.subRepo.GetByUserID(r.Context(), userCtx.UserID)
	if err != nil {
		respondError(w, http.StatusNotFound, "No active subscription found")
		return
	}

	// Cancel in Stripe
	_, err = h.paymentService.CancelSubscription(sub.StripeSubscriptionID, req.Immediately)
	if err != nil {
		log.Error().Err(err).Msg("Failed to cancel subscription")
		respondError(w, http.StatusInternalServerError, "Failed to cancel subscription")
		return
	}

	// Update in database
	if req.Immediately {
		sub.Status = "canceled"
		now := time.Now()
		sub.CanceledAt = &now
	} else {
		sub.CancelAtPeriodEnd = true
	}

	if err := h.subRepo.Update(r.Context(), sub); err != nil {
		log.Error().Err(err).Msg("Failed to update subscription")
	}

	respondJSON(w, http.StatusOK, sub)
}

// ReactivateSubscription reactivates a canceled subscription
func (h *BillingHandler) ReactivateSubscription(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sub, err := h.subRepo.GetByUserID(r.Context(), userCtx.UserID)
	if err != nil {
		respondError(w, http.StatusNotFound, "No subscription found")
		return
	}

	if !sub.CancelAtPeriodEnd {
		respondError(w, http.StatusBadRequest, "Subscription is not scheduled for cancellation")
		return
	}

	// Reactivate in Stripe
	_, err = h.paymentService.ReactivateSubscription(sub.StripeSubscriptionID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to reactivate subscription")
		respondError(w, http.StatusInternalServerError, "Failed to reactivate subscription")
		return
	}

	// Update in database
	sub.CancelAtPeriodEnd = false
	if err := h.subRepo.Update(r.Context(), sub); err != nil {
		log.Error().Err(err).Msg("Failed to update subscription")
	}

	respondJSON(w, http.StatusOK, sub)
}

type UpdateSubscriptionRequest struct {
	PriceID string `json:"price_id" validate:"required"`
}

// UpdateSubscription updates the subscription plan
func (h *BillingHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	sub, err := h.subRepo.GetByUserID(r.Context(), userCtx.UserID)
	if err != nil {
		respondError(w, http.StatusNotFound, "No active subscription found")
		return
	}

	// Update in Stripe
	_, err = h.paymentService.UpdateSubscription(sub.StripeSubscriptionID, req.PriceID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update subscription")
		respondError(w, http.StatusInternalServerError, "Failed to update subscription")
		return
	}

	// Update in database
	sub.StripePriceID = req.PriceID
	if err := h.subRepo.Update(r.Context(), sub); err != nil {
		log.Error().Err(err).Msg("Failed to update subscription in database")
	}

	respondJSON(w, http.StatusOK, sub)
}

// CreatePortalSession creates a Stripe customer portal session
func (h *BillingHandler) CreatePortalSession(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sub, err := h.subRepo.GetByUserID(r.Context(), userCtx.UserID)
	if err != nil {
		respondError(w, http.StatusNotFound, "No active subscription found")
		return
	}

	// Create portal session
	url, err := h.paymentService.CreateCustomerPortalSession(
		sub.StripeCustomerID,
		h.cfg.Server.FrontendURL+"/billing",
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create portal session")
		respondError(w, http.StatusInternalServerError, "Failed to create portal session")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"url": url})
}

// HandleWebhook handles Stripe webhook events
func (h *BillingHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read webhook body")
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Verify webhook signature
	event, err := webhook.ConstructEvent(body, r.Header.Get("Stripe-Signature"), h.cfg.Stripe.WebhookSecret)
	if err != nil {
		log.Error().Err(err).Msg("Failed to verify webhook signature")
		respondError(w, http.StatusBadRequest, "Invalid signature")
		return
	}

	// Handle different event types
	switch event.Type {
	case "checkout.session.completed":
		h.handleCheckoutSessionCompleted(event)
	case "customer.subscription.created":
		h.handleSubscriptionCreated(event)
	case "customer.subscription.updated":
		h.handleSubscriptionUpdated(event)
	case "customer.subscription.deleted":
		h.handleSubscriptionDeleted(event)
	case "invoice.paid":
		h.handleInvoicePaid(event)
	case "invoice.payment_failed":
		h.handleInvoicePaymentFailed(event)
	default:
		log.Info().Str("type", string(event.Type)).Msg("Unhandled webhook event type")
	}

	respondJSON(w, http.StatusOK, map[string]string{"received": "true"})
}

func (h *BillingHandler) handleCheckoutSessionCompleted(event stripe.Event) {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal checkout session")
		return
	}

	log.Info().
		Str("session_id", session.ID).
		Str("customer_id", session.Customer.ID).
		Str("subscription_id", session.Subscription.ID).
		Msg("Checkout session completed")
}

func (h *BillingHandler) handleSubscriptionCreated(event stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal subscription")
		return
	}

	// Get user ID from metadata (set during checkout)
	userID := sub.Metadata["user_id"]
	if userID == "" {
		log.Error().Msg("No user_id in subscription metadata")
		return
	}

	// Create subscription record
	subscription := &models.Subscription{
		UserID:               userID,
		StripeSubscriptionID: sub.ID,
		StripeCustomerID:     sub.Customer.ID,
		StripePriceID:        sub.Items.Data[0].Price.ID,
		Status:               string(sub.Status),
		CurrentPeriodStart:   time.Unix(sub.CurrentPeriodStart, 0),
		CurrentPeriodEnd:     time.Unix(sub.CurrentPeriodEnd, 0),
		CancelAtPeriodEnd:    sub.CancelAtPeriodEnd,
	}

	if sub.TrialStart != nil {
		start := time.Unix(*sub.TrialStart, 0)
		subscription.TrialStart = &start
	}

	if sub.TrialEnd != nil {
		end := time.Unix(*sub.TrialEnd, 0)
		subscription.TrialEnd = &end
	}

	ctx := context.Background()
	if err := h.subRepo.Create(ctx, subscription); err != nil {
		log.Error().Err(err).Msg("Failed to create subscription record")
	}
}

func (h *BillingHandler) handleSubscriptionUpdated(event stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal subscription")
		return
	}

	ctx := context.Background()
	// Get existing subscription
	subscription, err := h.subRepo.GetByStripeSubscriptionID(ctx, sub.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subscription")
		return
	}

	// Update subscription
	subscription.Status = string(sub.Status)
	subscription.CurrentPeriodStart = time.Unix(sub.CurrentPeriodStart, 0)
	subscription.CurrentPeriodEnd = time.Unix(sub.CurrentPeriodEnd, 0)
	subscription.CancelAtPeriodEnd = sub.CancelAtPeriodEnd

	if len(sub.Items.Data) > 0 {
		subscription.StripePriceID = sub.Items.Data[0].Price.ID
	}

	if err := h.subRepo.Update(ctx, subscription); err != nil {
		log.Error().Err(err).Msg("Failed to update subscription")
	}
}

func (h *BillingHandler) handleSubscriptionDeleted(event stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal subscription")
		return
	}

	ctx := context.Background()
	// Get existing subscription
	subscription, err := h.subRepo.GetByStripeSubscriptionID(ctx, sub.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subscription")
		return
	}

	// Mark as canceled
	if err := h.subRepo.Cancel(ctx, subscription.ID); err != nil {
		log.Error().Err(err).Msg("Failed to cancel subscription")
	}
}

func (h *BillingHandler) handleInvoicePaid(event stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal invoice")
		return
	}

	log.Info().
		Str("invoice_id", invoice.ID).
		Str("customer_id", invoice.Customer.ID).
		Int64("amount_paid", invoice.AmountPaid).
		Msg("Invoice paid")
}

func (h *BillingHandler) handleInvoicePaymentFailed(event stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal invoice")
		return
	}

	log.Warn().
		Str("invoice_id", invoice.ID).
		Str("customer_id", invoice.Customer.ID).
		Msg("Invoice payment failed")
}
