package models

import "time"

type Subscription struct {
	ID                   string     `json:"id"`
	UserID               string     `json:"user_id"`
	StripeSubscriptionID string     `json:"stripe_subscription_id"`
	StripeCustomerID     string     `json:"stripe_customer_id"`
	StripePriceID        string     `json:"stripe_price_id"`
	Status               string     `json:"status"` // active, trialing, past_due, canceled, unpaid
	CurrentPeriodStart   time.Time  `json:"current_period_start"`
	CurrentPeriodEnd     time.Time  `json:"current_period_end"`
	CancelAtPeriodEnd    bool       `json:"cancel_at_period_end"`
	TrialStart           *time.Time `json:"trial_start,omitempty"`
	TrialEnd             *time.Time `json:"trial_end,omitempty"`
	CanceledAt           *time.Time `json:"canceled_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type Invoice struct {
	ID                   string     `json:"id"`
	UserID               string     `json:"user_id"`
	SubscriptionID       *string    `json:"subscription_id,omitempty"`
	StripeInvoiceID      string     `json:"stripe_invoice_id"`
	StripePaymentIntentID *string   `json:"stripe_payment_intent_id,omitempty"`
	AmountDue            int64      `json:"amount_due"`
	AmountPaid           int64      `json:"amount_paid"`
	Currency             string     `json:"currency"`
	Status               string     `json:"status"` // draft, open, paid, uncollectible, void
	InvoicePDF           *string    `json:"invoice_pdf,omitempty"`
	HostedInvoiceURL     *string    `json:"hosted_invoice_url,omitempty"`
	DueDate              *time.Time `json:"due_date,omitempty"`
	PaidAt               *time.Time `json:"paid_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}
