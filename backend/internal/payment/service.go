package payment

import (
	"fmt"

	"github.com/mahendrakalkura/saas-blueprint/internal/config"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/paymentmethod"
	"github.com/stripe/stripe-go/v81/price"
	"github.com/stripe/stripe-go/v81/subscription"
	"github.com/stripe/stripe-go/v81/webhookendpoint"
)

type Service struct {
	secretKey string
}

type CreateCheckoutSessionParams struct {
	CustomerEmail string
	PriceID       string
	SuccessURL    string
	CancelURL     string
	TrialDays     int64
	Metadata      map[string]string
}

type SubscriptionInfo struct {
	ID                 string
	CustomerID         string
	Status             string
	CurrentPeriodStart int64
	CurrentPeriodEnd   int64
	CancelAtPeriodEnd  bool
	TrialStart         *int64
	TrialEnd           *int64
	PriceID            string
	ProductID          string
}

func NewService(cfg *config.StripeConfig) *Service {
	stripe.Key = cfg.SecretKey
	return &Service{
		secretKey: cfg.SecretKey,
	}
}

// CreateCheckoutSession creates a new Stripe Checkout session for subscription
func (s *Service) CreateCheckoutSession(params CreateCheckoutSessionParams) (string, error) {
	sessionParams := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(params.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(params.SuccessURL),
		CancelURL:  stripe.String(params.CancelURL),
	}

	if params.CustomerEmail != "" {
		sessionParams.CustomerEmail = stripe.String(params.CustomerEmail)
	}

	if params.TrialDays > 0 {
		sessionParams.SubscriptionData = &stripe.CheckoutSessionSubscriptionDataParams{
			TrialPeriodDays: stripe.Int64(params.TrialDays),
		}
	}

	if len(params.Metadata) > 0 {
		sessionParams.Metadata = params.Metadata
	}

	sess, err := session.New(sessionParams)
	if err != nil {
		return "", fmt.Errorf("failed to create checkout session: %w", err)
	}

	return sess.URL, nil
}

// CreateCustomer creates a new Stripe customer
func (s *Service) CreateCustomer(email, name string, metadata map[string]string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
	}

	if len(metadata) > 0 {
		params.Metadata = metadata
	}

	cust, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return cust, nil
}

// GetCustomer retrieves a Stripe customer
func (s *Service) GetCustomer(customerID string) (*stripe.Customer, error) {
	cust, err := customer.Get(customerID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return cust, nil
}

// UpdateCustomer updates a Stripe customer
func (s *Service) UpdateCustomer(customerID string, params *stripe.CustomerParams) (*stripe.Customer, error) {
	cust, err := customer.Update(customerID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	return cust, nil
}

// CreateSubscription creates a new subscription for a customer
func (s *Service) CreateSubscription(customerID, priceID string, trialDays int64, metadata map[string]string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(priceID),
			},
		},
	}

	if trialDays > 0 {
		params.TrialPeriodDays = stripe.Int64(trialDays)
	}

	if len(metadata) > 0 {
		params.Metadata = metadata
	}

	sub, err := subscription.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return sub, nil
}

// GetSubscription retrieves a subscription
func (s *Service) GetSubscription(subscriptionID string) (*SubscriptionInfo, error) {
	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	info := &SubscriptionInfo{
		ID:                 sub.ID,
		CustomerID:         sub.Customer.ID,
		Status:             string(sub.Status),
		CurrentPeriodStart: sub.CurrentPeriodStart,
		CurrentPeriodEnd:   sub.CurrentPeriodEnd,
		CancelAtPeriodEnd:  sub.CancelAtPeriodEnd,
	}

	if sub.TrialStart != nil {
		start := *sub.TrialStart
		info.TrialStart = &start
	}

	if sub.TrialEnd != nil {
		end := *sub.TrialEnd
		info.TrialEnd = &end
	}

	if len(sub.Items.Data) > 0 {
		info.PriceID = sub.Items.Data[0].Price.ID
		info.ProductID = sub.Items.Data[0].Price.Product.ID
	}

	return info, nil
}

// CancelSubscription cancels a subscription at period end
func (s *Service) CancelSubscription(subscriptionID string, immediately bool) (*stripe.Subscription, error) {
	var sub *stripe.Subscription
	var err error

	if immediately {
		sub, err = subscription.Cancel(subscriptionID, nil)
	} else {
		params := &stripe.SubscriptionParams{
			CancelAtPeriodEnd: stripe.Bool(true),
		}
		sub, err = subscription.Update(subscriptionID, params)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to cancel subscription: %w", err)
	}

	return sub, nil
}

// ReactivateSubscription reactivates a canceled subscription
func (s *Service) ReactivateSubscription(subscriptionID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(false),
	}

	sub, err := subscription.Update(subscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to reactivate subscription: %w", err)
	}

	return sub, nil
}

// UpdateSubscription updates a subscription (e.g., change plan)
func (s *Service) UpdateSubscription(subscriptionID, newPriceID string) (*stripe.Subscription, error) {
	// Get current subscription
	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	if len(sub.Items.Data) == 0 {
		return nil, fmt.Errorf("subscription has no items")
	}

	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(sub.Items.Data[0].ID),
				Price: stripe.String(newPriceID),
			},
		},
	}

	updated, err := subscription.Update(subscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	return updated, nil
}

// ListPaymentMethods lists payment methods for a customer
func (s *Service) ListPaymentMethods(customerID string) ([]*stripe.PaymentMethod, error) {
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String("card"),
	}

	i := paymentmethod.List(params)
	methods := []*stripe.PaymentMethod{}

	for i.Next() {
		methods = append(methods, i.PaymentMethod())
	}

	if err := i.Err(); err != nil {
		return nil, fmt.Errorf("failed to list payment methods: %w", err)
	}

	return methods, nil
}

// DetachPaymentMethod removes a payment method
func (s *Service) DetachPaymentMethod(paymentMethodID string) error {
	_, err := paymentmethod.Detach(paymentMethodID, nil)
	if err != nil {
		return fmt.Errorf("failed to detach payment method: %w", err)
	}

	return nil
}

// CreateCustomerPortalSession creates a customer portal session
func (s *Service) CreateCustomerPortalSession(customerID, returnURL string) (string, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create portal session: %w", err)
	}

	return sess.URL, nil
}

// ListPrices lists all active prices for a product
func (s *Service) ListPrices(productID string) ([]*stripe.Price, error) {
	params := &stripe.PriceListParams{
		Product: stripe.String(productID),
		Active:  stripe.Bool(true),
	}

	i := price.List(params)
	prices := []*stripe.Price{}

	for i.Next() {
		prices = append(prices, i.Price())
	}

	if err := i.Err(); err != nil {
		return nil, fmt.Errorf("failed to list prices: %w", err)
	}

	return prices, nil
}
