package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mahendrakalkura/saas-blueprint/internal/models"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	query := `
		INSERT INTO notifications (id, user_id, type, title, message, metadata, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`

	notification.ID = uuid.New().String()
	notification.CreatedAt = time.Now()
	notification.IsRead = false

	// Convert metadata map to JSON
	var metadataJSON []byte
	var err error
	if notification.Metadata != nil {
		metadataJSON, err = json.Marshal(notification.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal notification metadata: %w", err)
		}
	}

	err = r.db.QueryRowContext(ctx, query,
		notification.ID,
		notification.UserID,
		notification.Type,
		notification.Title,
		notification.Message,
		metadataJSON,
		notification.IsRead,
		notification.CreatedAt,
	).Scan(&notification.ID, &notification.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

func (r *NotificationRepository) GetByID(ctx context.Context, id string) (*models.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, metadata, is_read, read_at, created_at
		FROM notifications
		WHERE id = $1
	`

	notification := &models.Notification{}
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&notification.Title,
		&notification.Message,
		&metadataJSON,
		&notification.IsRead,
		&notification.ReadAt,
		&notification.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("notification not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	// Unmarshal metadata JSON
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &notification.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal notification metadata: %w", err)
		}
	}

	return notification, nil
}

func (r *NotificationRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, metadata, is_read, read_at, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	notifications := []*models.Notification{}
	for rows.Next() {
		notification := &models.Notification{}
		var metadataJSON []byte

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&metadataJSON,
			&notification.IsRead,
			&notification.ReadAt,
			&notification.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		// Unmarshal metadata JSON
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &notification.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal notification metadata: %w", err)
			}
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, id string) error {
	query := `
		UPDATE notifications
		SET is_read = true, read_at = $1
		WHERE id = $2 AND is_read = false
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("notification not found or already read")
	}

	return nil
}

func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	query := `
		UPDATE notifications
		SET is_read = true, read_at = $1
		WHERE user_id = $2 AND is_read = false
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, now, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

func (r *NotificationRepository) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1 AND is_read = false
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

func (r *NotificationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM notifications WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}
