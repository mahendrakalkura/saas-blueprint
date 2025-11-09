package email

import (
	"fmt"

	"github.com/mahendrakalkura/saas-blueprint/internal/config"
	"github.com/resend/resend-go/v2"
)

type Service struct {
	client   *resend.Client
	fromEmail string
	fromName  string
}

func NewService(cfg *config.EmailConfig) *Service {
	client := resend.NewClient(cfg.APIKey)

	return &Service{
		client:    client,
		fromEmail: cfg.FromEmail,
		fromName:  cfg.FromName,
	}
}

type SendEmailParams struct {
	To      string
	Subject string
	HTML    string
}

func (s *Service) SendEmail(params SendEmailParams) error {
	if s.client == nil {
		return fmt.Errorf("email service not configured")
	}

	req := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail),
		To:      []string{params.To},
		Subject: params.Subject,
		Html:    params.HTML,
	}

	_, err := s.client.Emails.Send(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *Service) SendWelcomeEmail(to, firstName string) error {
	html := renderWelcomeEmail(firstName)
	return s.SendEmail(SendEmailParams{
		To:      to,
		Subject: "Welcome to SaaS Blueprint!",
		HTML:    html,
	})
}

func (s *Service) SendVerificationEmail(to, token string, baseURL string) error {
	html := renderVerificationEmail(token, baseURL)
	return s.SendEmail(SendEmailParams{
		To:      to,
		Subject: "Verify your email address",
		HTML:    html,
	})
}

func (s *Service) SendPasswordResetEmail(to, token string, baseURL string) error {
	html := renderPasswordResetEmail(token, baseURL)
	return s.SendEmail(SendEmailParams{
		To:      to,
		Subject: "Reset your password",
		HTML:    html,
	})
}
