package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mahendrakalkura/saas-blueprint/internal/models"
)

type SubscriptionRepository struct {
	db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *models.Subscription) error {
	query := `
		INSERT INTO subscriptions (
			id, user_id, stripe_subscription_id, stripe_customer_id, stripe_price_id,
			status, current_period_start, current_period_end, cancel_at_period_end,
			trial_start, trial_end, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`

	sub.ID = uuid.New().String()
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, query,
		sub.ID,
		sub.UserID,
		sub.StripeSubscriptionID,
		sub.StripeCustomerID,
		sub.StripePriceID,
		sub.Status,
		sub.CurrentPeriodStart,
		sub.CurrentPeriodEnd,
		sub.CancelAtPeriodEnd,
		sub.TrialStart,
		sub.TrialEnd,
		sub.CreatedAt,
		sub.UpdatedAt,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id string) (*models.Subscription, error) {
	query := `
		SELECT id, user_id, stripe_subscription_id, stripe_customer_id, stripe_price_id,
		       status, current_period_start, current_period_end, cancel_at_period_end,
		       trial_start, trial_end, canceled_at, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`

	sub := &models.Subscription{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sub.ID,
		&sub.UserID,
		&sub.StripeSubscriptionID,
		&sub.StripeCustomerID,
		&sub.StripePriceID,
		&sub.Status,
		&sub.CurrentPeriodStart,
		&sub.CurrentPeriodEnd,
		&sub.CancelAtPeriodEnd,
		&sub.TrialStart,
		&sub.TrialEnd,
		&sub.CanceledAt,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("subscription not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return sub, nil
}

func (r *SubscriptionRepository) GetByStripeSubscriptionID(ctx context.Context, stripeSubID string) (*models.Subscription, error) {
	query := `
		SELECT id, user_id, stripe_subscription_id, stripe_customer_id, stripe_price_id,
		       status, current_period_start, current_period_end, cancel_at_period_end,
		       trial_start, trial_end, canceled_at, created_at, updated_at
		FROM subscriptions
		WHERE stripe_subscription_id = $1
	`

	sub := &models.Subscription{}
	err := r.db.QueryRowContext(ctx, query, stripeSubID).Scan(
		&sub.ID,
		&sub.UserID,
		&sub.StripeSubscriptionID,
		&sub.StripeCustomerID,
		&sub.StripePriceID,
		&sub.Status,
		&sub.CurrentPeriodStart,
		&sub.CurrentPeriodEnd,
		&sub.CancelAtPeriodEnd,
		&sub.TrialStart,
		&sub.TrialEnd,
		&sub.CanceledAt,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("subscription not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return sub, nil
}

func (r *SubscriptionRepository) GetByUserID(ctx context.Context, userID string) (*models.Subscription, error) {
	query := `
		SELECT id, user_id, stripe_subscription_id, stripe_customer_id, stripe_price_id,
		       status, current_period_start, current_period_end, cancel_at_period_end,
		       trial_start, trial_end, canceled_at, created_at, updated_at
		FROM subscriptions
		WHERE user_id = $1 AND status IN ('active', 'trialing')
		ORDER BY created_at DESC
		LIMIT 1
	`

	sub := &models.Subscription{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&sub.ID,
		&sub.UserID,
		&sub.StripeSubscriptionID,
		&sub.StripeCustomerID,
		&sub.StripePriceID,
		&sub.Status,
		&sub.CurrentPeriodStart,
		&sub.CurrentPeriodEnd,
		&sub.CancelAtPeriodEnd,
		&sub.TrialStart,
		&sub.TrialEnd,
		&sub.CanceledAt,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("subscription not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return sub, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, sub *models.Subscription) error {
	query := `
		UPDATE subscriptions
		SET stripe_price_id = $1, status = $2, current_period_start = $3,
		    current_period_end = $4, cancel_at_period_end = $5, trial_start = $6,
		    trial_end = $7, canceled_at = $8, updated_at = $9
		WHERE id = $10
	`

	sub.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		sub.StripePriceID,
		sub.Status,
		sub.CurrentPeriodStart,
		sub.CurrentPeriodEnd,
		sub.CancelAtPeriodEnd,
		sub.TrialStart,
		sub.TrialEnd,
		sub.CanceledAt,
		sub.UpdatedAt,
		sub.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("subscription not found")
	}

	return nil
}

func (r *SubscriptionRepository) Cancel(ctx context.Context, id string) error {
	query := `
		UPDATE subscriptions
		SET status = 'canceled', canceled_at = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("subscription not found")
	}

	return nil
}
