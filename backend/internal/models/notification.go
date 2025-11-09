package models

import "time"

type Notification struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Type      string                 `json:"type"` // organization_invite, member_added, etc.
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	IsRead    bool                   `json:"is_read"`
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
