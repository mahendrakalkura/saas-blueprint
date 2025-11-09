package notification

import (
	"context"

	"github.com/mahendrakalkura/saas-blueprint/internal/models"
	"github.com/mahendrakalkura/saas-blueprint/internal/repository"
	"github.com/mahendrakalkura/saas-blueprint/internal/websocket"
	"github.com/rs/zerolog/log"
)

// Service handles notification creation and delivery
type Service struct {
	notificationRepo *repository.NotificationRepository
	wsHub            *websocket.Hub
}

// NewService creates a new notification service
func NewService(notificationRepo *repository.NotificationRepository, wsHub *websocket.Hub) *Service {
	return &Service{
		notificationRepo: notificationRepo,
		wsHub:            wsHub,
	}
}

// CreateAndSend creates a notification in the database and sends it via WebSocket
func (s *Service) CreateAndSend(ctx context.Context, notification *models.Notification) error {
	// Save to database
	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		log.Error().Err(err).Msg("Failed to create notification in database")
		return err
	}

	// Send via WebSocket
	if err := s.wsHub.SendToUser(notification.UserID, "notification", map[string]interface{}{
		"id":         notification.ID,
		"type":       notification.Type,
		"title":      notification.Title,
		"message":    notification.Message,
		"data":       notification.Data,
		"read":       notification.Read,
		"created_at": notification.CreatedAt,
	}); err != nil {
		log.Error().Err(err).Msg("Failed to send notification via WebSocket")
		// Don't return error - notification is already saved
	}

	log.Info().
		Str("notification_id", notification.ID).
		Str("user_id", notification.UserID).
		Str("type", notification.Type).
		Msg("Notification created and sent")

	return nil
}

// SendWelcomeNotification sends a welcome notification to a new user
func (s *Service) SendWelcomeNotification(ctx context.Context, userID, userName string) error {
	notification := &models.Notification{
		UserID:  userID,
		Type:    models.NotificationTypeWelcome,
		Title:   "Welcome!",
		Message: "Welcome to our platform, " + userName + "! We're excited to have you here.",
		Data: map[string]interface{}{
			"user_name": userName,
		},
	}

	return s.CreateAndSend(ctx, notification)
}

// SendOrganizationInviteNotification sends a notification when a user is invited to an organization
func (s *Service) SendOrganizationInviteNotification(ctx context.Context, userID, orgName, inviterName string) error {
	notification := &models.Notification{
		UserID:  userID,
		Type:    models.NotificationTypeOrganizationInvite,
		Title:   "Organization Invitation",
		Message: inviterName + " invited you to join " + orgName,
		Data: map[string]interface{}{
			"organization_name": orgName,
			"inviter_name":      inviterName,
		},
	}

	return s.CreateAndSend(ctx, notification)
}

// SendMemberAddedNotification sends a notification when a user is added to an organization
func (s *Service) SendMemberAddedNotification(ctx context.Context, userID, orgName, role string) error {
	notification := &models.Notification{
		UserID:  userID,
		Type:    models.NotificationTypeMemberAdded,
		Title:   "Added to Organization",
		Message: "You've been added to " + orgName + " as " + role,
		Data: map[string]interface{}{
			"organization_name": orgName,
			"role":              role,
		},
	}

	return s.CreateAndSend(ctx, notification)
}

// SendRoleChangedNotification sends a notification when a user's role changes
func (s *Service) SendRoleChangedNotification(ctx context.Context, userID, orgName, oldRole, newRole string) error {
	notification := &models.Notification{
		UserID:  userID,
		Type:    models.NotificationTypeRoleChanged,
		Title:   "Role Changed",
		Message: "Your role in " + orgName + " has been changed from " + oldRole + " to " + newRole,
		Data: map[string]interface{}{
			"organization_name": orgName,
			"old_role":          oldRole,
			"new_role":          newRole,
		},
	}

	return s.CreateAndSend(ctx, notification)
}
