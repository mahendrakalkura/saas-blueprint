package models

import "time"

type Notification struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Type      string                 `json:"type"` // organization_invite, member_added, etc.
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Read      bool                   `json:"read"`
	ReadAt    *time.Time             `json:"read_at,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// NotificationType constants
const (
	NotificationTypeOrganizationInvite = "organization_invite"
	NotificationTypeMemberAdded        = "member_added"
	NotificationTypeMemberRemoved      = "member_removed"
	NotificationTypeRoleChanged        = "role_changed"
	NotificationTypeWelcome            = "welcome"
)
