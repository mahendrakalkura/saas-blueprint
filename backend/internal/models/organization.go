package models

import "time"

type Organization struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	OwnerID   string     `json:"owner_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}

type OrganizationMember struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	UserID         string     `json:"user_id"`
	Role           string     `json:"role"` // owner, admin, member
	InvitedBy      *string    `json:"invited_by,omitempty"`
	InvitedAt      *time.Time `json:"invited_at,omitempty"`
	JoinedAt       *time.Time `json:"joined_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type OrganizationInvitation struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	Email          string     `json:"email"`
	Role           string     `json:"role"`
	Token          string     `json:"-"`
	InvitedBy      string     `json:"invited_by"`
	ExpiresAt      time.Time  `json:"expires_at"`
	AcceptedAt     *time.Time `json:"accepted_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// MemberWithUser combines member info with user details
type MemberWithUser struct {
	Member *OrganizationMember `json:"member"`
	User   *User               `json:"user"`
}
