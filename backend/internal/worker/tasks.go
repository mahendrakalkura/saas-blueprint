package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/mahendrakalkura/saas-blueprint/internal/email"
	"github.com/rs/zerolog"
)

const (
	TypeEmailWelcome      = "email:welcome"
	TypeEmailVerification = "email:verification"
	TypeEmailPasswordReset = "email:password_reset"
	TypeCleanupSessions   = "cleanup:sessions"
)

type EmailWelcomePayload struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
}

type EmailVerificationPayload struct {
	Email   string `json:"email"`
	Token   string `json:"token"`
	BaseURL string `json:"base_url"`
}

type EmailPasswordResetPayload struct {
	Email   string `json:"email"`
	Token   string `json:"token"`
	BaseURL string `json:"base_url"`
}

type TaskProcessor struct {
	emailService *email.Service
	sessionRepo  SessionRepository
	logger       *zerolog.Logger
}

// SessionRepository interface for session cleanup
type SessionRepository interface {
	DeleteExpired(ctx context.Context) error
}

func NewTaskProcessor(emailService *email.Service, sessionRepo SessionRepository, logger *zerolog.Logger) *TaskProcessor {
	return &TaskProcessor{
		emailService: emailService,
		sessionRepo:  sessionRepo,
		logger:       logger,
	}
}

// ProcessEmailWelcome sends a welcome email
func (p *TaskProcessor) ProcessEmailWelcome(ctx context.Context, t *asynq.Task) error {
	var payload EmailWelcomePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	p.logger.Info().
		Str("task", TypeEmailWelcome).
		Str("email", payload.Email).
		Msg("Processing welcome email")

	if err := p.emailService.SendWelcomeEmail(payload.Email, payload.FirstName); err != nil {
		p.logger.Error().Err(err).Str("email", payload.Email).Msg("Failed to send welcome email")
		return fmt.Errorf("failed to send welcome email: %w", err)
	}

	p.logger.Info().Str("email", payload.Email).Msg("Welcome email sent successfully")
	return nil
}

// ProcessEmailVerification sends an email verification email
func (p *TaskProcessor) ProcessEmailVerification(ctx context.Context, t *asynq.Task) error {
	var payload EmailVerificationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	p.logger.Info().
		Str("task", TypeEmailVerification).
		Str("email", payload.Email).
		Msg("Processing verification email")

	if err := p.emailService.SendVerificationEmail(payload.Email, payload.Token, payload.BaseURL); err != nil {
		p.logger.Error().Err(err).Str("email", payload.Email).Msg("Failed to send verification email")
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	p.logger.Info().Str("email", payload.Email).Msg("Verification email sent successfully")
	return nil
}

// ProcessEmailPasswordReset sends a password reset email
func (p *TaskProcessor) ProcessEmailPasswordReset(ctx context.Context, t *asynq.Task) error {
	var payload EmailPasswordResetPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	p.logger.Info().
		Str("task", TypeEmailPasswordReset).
		Str("email", payload.Email).
		Msg("Processing password reset email")

	if err := p.emailService.SendPasswordResetEmail(payload.Email, payload.Token, payload.BaseURL); err != nil {
		p.logger.Error().Err(err).Str("email", payload.Email).Msg("Failed to send password reset email")
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	p.logger.Info().Str("email", payload.Email).Msg("Password reset email sent successfully")
	return nil
}

// ProcessCleanupSessions removes expired sessions from database
func (p *TaskProcessor) ProcessCleanupSessions(ctx context.Context, t *asynq.Task) error {
	p.logger.Info().Str("task", TypeCleanupSessions).Msg("Cleaning up expired sessions")

	if err := p.sessionRepo.DeleteExpired(ctx); err != nil {
		p.logger.Error().Err(err).Msg("Failed to delete expired sessions")
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	p.logger.Info().Msg("Session cleanup completed successfully")
	return nil
}
