package worker

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

type Client struct {
	client *asynq.Client
}

func NewClient(redisAddr string) *Client {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return &Client{client: client}
}

func (c *Client) Close() error {
	return c.client.Close()
}

// EnqueueWelcomeEmail enqueues a welcome email task
func (c *Client) EnqueueWelcomeEmail(email, firstName string) error {
	payload, err := json.Marshal(EmailWelcomePayload{
		Email:     email,
		FirstName: firstName,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeEmailWelcome, payload)
	_, err = c.client.Enqueue(task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

// EnqueueVerificationEmail enqueues an email verification task
func (c *Client) EnqueueVerificationEmail(email, token, baseURL string) error {
	payload, err := json.Marshal(EmailVerificationPayload{
		Email:   email,
		Token:   token,
		BaseURL: baseURL,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeEmailVerification, payload)
	_, err = c.client.Enqueue(task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

// EnqueuePasswordResetEmail enqueues a password reset email task
func (c *Client) EnqueuePasswordResetEmail(email, token, baseURL string) error {
	payload, err := json.Marshal(EmailPasswordResetPayload{
		Email:   email,
		Token:   token,
		BaseURL: baseURL,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeEmailPasswordReset, payload)
	_, err = c.client.Enqueue(task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

// ScheduleSessionCleanup schedules periodic session cleanup
func (c *Client) ScheduleSessionCleanup(cronspec string) error {
	task := asynq.NewTask(TypeCleanupSessions, nil)

	_, err := c.client.Enqueue(task, asynq.ProcessIn(10*time.Second))
	if err != nil {
		return fmt.Errorf("failed to schedule task: %w", err)
	}

	return nil
}
