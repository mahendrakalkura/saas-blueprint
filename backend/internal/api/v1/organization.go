package v1

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mahendrakalkura/saas-blueprint/internal/auth"
	"github.com/mahendrakalkura/saas-blueprint/internal/models"
	"github.com/mahendrakalkura/saas-blueprint/internal/repository"
	"github.com/mahendrakalkura/saas-blueprint/internal/worker"
	"github.com/rs/zerolog/log"
)

type OrganizationHandler struct {
	orgRepo    *repository.OrganizationRepository
	memberRepo *repository.OrganizationMemberRepository
	userRepo   *repository.UserRepository
	worker     *worker.Client
}

func NewOrganizationHandler(orgRepo *repository.OrganizationRepository, memberRepo *repository.OrganizationMemberRepository, userRepo *repository.UserRepository, worker *worker.Client) *OrganizationHandler {
	return &OrganizationHandler{
		orgRepo:    orgRepo,
		memberRepo: memberRepo,
		userRepo:   userRepo,
		worker:     worker,
	}
}

type CreateOrganizationRequest struct {
	Name string `json:"name" validate:"required,min=3,max=100"`
}

type UpdateOrganizationRequest struct {
	Name string `json:"name" validate:"required,min=3,max=100"`
}

type InviteMemberRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=admin member"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=owner admin member"`
}

// CreateOrganization creates a new organization
func (h *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Generate slug from name
	slug := generateSlug(req.Name)

	// Check if slug already exists
	existing, _ := h.orgRepo.GetBySlug(r.Context(), slug)
	if existing != nil {
		respondError(w, http.StatusConflict, "Organization with this name already exists")
		return
	}

	// Create organization
	org := &models.Organization{
		Name:    req.Name,
		Slug:    slug,
		OwnerID: userCtx.UserID,
	}

	if err := h.orgRepo.Create(r.Context(), org); err != nil {
		log.Error().Err(err).Msg("Failed to create organization")
		respondError(w, http.StatusInternalServerError, "Failed to create organization")
		return
	}

	// Add creator as owner member
	now := time.Now()
	member := &models.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         userCtx.UserID,
		Role:           "owner",
		JoinedAt:       &now,
	}

	if err := h.memberRepo.Create(r.Context(), member); err != nil {
		log.Error().Err(err).Msg("Failed to add owner as member")
		// Don't fail the request, just log the error
	}

	respondJSON(w, http.StatusCreated, org)
}

// ListOrganizations lists all organizations the user is a member of
func (h *OrganizationHandler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orgs, err := h.orgRepo.ListByUserID(r.Context(), userCtx.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list organizations")
		respondError(w, http.StatusInternalServerError, "Failed to list organizations")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"organizations": orgs,
	})
}

// GetOrganization retrieves a single organization
func (h *OrganizationHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orgID := chi.URLParam(r, "id")
	if orgID == "" {
		respondError(w, http.StatusBadRequest, "Organization ID is required")
		return
	}

	org, err := h.orgRepo.GetByID(r.Context(), orgID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Organization not found")
		return
	}

	// Check if user is a member
	_, err = h.memberRepo.GetByOrganizationAndUser(r.Context(), orgID, userCtx.UserID)
	if err != nil {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	respondJSON(w, http.StatusOK, org)
}

// UpdateOrganization updates an organization
func (h *OrganizationHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orgID := chi.URLParam(r, "id")
	if orgID == "" {
		respondError(w, http.StatusBadRequest, "Organization ID is required")
		return
	}

	var req UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Check if user is owner or admin
	member, err := h.memberRepo.GetByOrganizationAndUser(r.Context(), orgID, userCtx.UserID)
	if err != nil || (member.Role != "owner" && member.Role != "admin") {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	org, err := h.orgRepo.GetByID(r.Context(), orgID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Organization not found")
		return
	}

	// Update fields
	org.Name = req.Name
	org.Slug = generateSlug(req.Name)

	if err := h.orgRepo.Update(r.Context(), org); err != nil {
		log.Error().Err(err).Msg("Failed to update organization")
		respondError(w, http.StatusInternalServerError, "Failed to update organization")
		return
	}

	respondJSON(w, http.StatusOK, org)
}

// DeleteOrganization deletes an organization
func (h *OrganizationHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orgID := chi.URLParam(r, "id")
	if orgID == "" {
		respondError(w, http.StatusBadRequest, "Organization ID is required")
		return
	}

	// Check if user is owner
	org, err := h.orgRepo.GetByID(r.Context(), orgID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Organization not found")
		return
	}

	if org.OwnerID != userCtx.UserID {
		respondError(w, http.StatusForbidden, "Only the owner can delete the organization")
		return
	}

	if err := h.orgRepo.Delete(r.Context(), orgID); err != nil {
		log.Error().Err(err).Msg("Failed to delete organization")
		respondError(w, http.StatusInternalServerError, "Failed to delete organization")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Organization deleted successfully"})
}

// ListMembers lists all members of an organization
func (h *OrganizationHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orgID := chi.URLParam(r, "id")
	if orgID == "" {
		respondError(w, http.StatusBadRequest, "Organization ID is required")
		return
	}

	// Check if user is a member
	_, err := h.memberRepo.GetByOrganizationAndUser(r.Context(), orgID, userCtx.UserID)
	if err != nil {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	members, err := h.memberRepo.ListByOrganization(r.Context(), orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list members")
		respondError(w, http.StatusInternalServerError, "Failed to list members")
		return
	}

	// Fetch user details for each member
	membersWithUsers := make([]map[string]interface{}, 0, len(members))
	for _, member := range members {
		user, err := h.userRepo.GetByID(r.Context(), member.UserID)
		if err != nil {
			continue
		}

		membersWithUsers = append(membersWithUsers, map[string]interface{}{
			"member": member,
			"user":   user,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"members": membersWithUsers,
	})
}

// InviteMember invites a new member to the organization
func (h *OrganizationHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orgID := chi.URLParam(r, "id")
	if orgID == "" {
		respondError(w, http.StatusBadRequest, "Organization ID is required")
		return
	}

	var req InviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Check if user is owner or admin
	member, err := h.memberRepo.GetByOrganizationAndUser(r.Context(), orgID, userCtx.UserID)
	if err != nil || (member.Role != "owner" && member.Role != "admin") {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Generate invitation token
	token, err := generateToken()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate invitation token")
		return
	}

	// Create invitation
	invitation := &models.OrganizationInvitation{
		OrganizationID: orgID,
		Email:          strings.ToLower(strings.TrimSpace(req.Email)),
		Role:           req.Role,
		Token:          token,
		InvitedBy:      userCtx.UserID,
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := h.memberRepo.CreateInvitation(r.Context(), invitation); err != nil {
		log.Error().Err(err).Msg("Failed to create invitation")
		respondError(w, http.StatusInternalServerError, "Failed to create invitation")
		return
	}

	// TODO: Send invitation email via worker
	// h.worker.EnqueueInvitationEmail(invitation.Email, token, orgID)

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"invitation": invitation,
		"message":    "Invitation sent successfully",
	})
}

// AcceptInvitation accepts an organization invitation
func (h *OrganizationHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	token := chi.URLParam(r, "token")
	if token == "" {
		respondError(w, http.StatusBadRequest, "Invitation token is required")
		return
	}

	invitation, err := h.memberRepo.GetInvitationByToken(r.Context(), token)
	if err != nil {
		respondError(w, http.StatusNotFound, "Invitation not found or expired")
		return
	}

	// Check if invitation has expired
	if time.Now().After(invitation.ExpiresAt) {
		respondError(w, http.StatusBadRequest, "Invitation has expired")
		return
	}

	// Get user
	user, err := h.userRepo.GetByID(r.Context(), userCtx.UserID)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	// Check if email matches
	if strings.ToLower(user.Email) != strings.ToLower(invitation.Email) {
		respondError(w, http.StatusForbidden, "This invitation was sent to a different email address")
		return
	}

	// Check if already a member
	_, err = h.memberRepo.GetByOrganizationAndUser(r.Context(), invitation.OrganizationID, userCtx.UserID)
	if err == nil {
		respondError(w, http.StatusConflict, "Already a member of this organization")
		return
	}

	// Add as member
	now := time.Now()
	member := &models.OrganizationMember{
		OrganizationID: invitation.OrganizationID,
		UserID:         userCtx.UserID,
		Role:           invitation.Role,
		InvitedBy:      &invitation.InvitedBy,
		InvitedAt:      &invitation.CreatedAt,
		JoinedAt:       &now,
	}

	if err := h.memberRepo.Create(r.Context(), member); err != nil {
		log.Error().Err(err).Msg("Failed to add member")
		respondError(w, http.StatusInternalServerError, "Failed to join organization")
		return
	}

	// Mark invitation as accepted
	if err := h.memberRepo.AcceptInvitation(r.Context(), token); err != nil {
		log.Error().Err(err).Msg("Failed to mark invitation as accepted")
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"member":  member,
		"message": "Successfully joined organization",
	})
}

// UpdateMemberRole updates a member's role
func (h *OrganizationHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orgID := chi.URLParam(r, "id")
	memberID := chi.URLParam(r, "memberID")
	if orgID == "" || memberID == "" {
		respondError(w, http.StatusBadRequest, "Organization ID and Member ID are required")
		return
	}

	var req UpdateMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Check if user is owner
	org, err := h.orgRepo.GetByID(r.Context(), orgID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Organization not found")
		return
	}

	if org.OwnerID != userCtx.UserID {
		respondError(w, http.StatusForbidden, "Only the owner can update member roles")
		return
	}

	// Cannot change owner role
	targetMember, err := h.memberRepo.GetByID(r.Context(), memberID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Member not found")
		return
	}

	if targetMember.Role == "owner" {
		respondError(w, http.StatusBadRequest, "Cannot change owner's role")
		return
	}

	if err := h.memberRepo.UpdateRole(r.Context(), memberID, req.Role); err != nil {
		log.Error().Err(err).Msg("Failed to update member role")
		respondError(w, http.StatusInternalServerError, "Failed to update member role")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Member role updated successfully"})
}

// RemoveMember removes a member from the organization
func (h *OrganizationHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orgID := chi.URLParam(r, "id")
	memberID := chi.URLParam(r, "memberID")
	if orgID == "" || memberID == "" {
		respondError(w, http.StatusBadRequest, "Organization ID and Member ID are required")
		return
	}

	// Check if user is owner or admin
	currentMember, err := h.memberRepo.GetByOrganizationAndUser(r.Context(), orgID, userCtx.UserID)
	if err != nil || (currentMember.Role != "owner" && currentMember.Role != "admin") {
		respondError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Cannot remove owner
	targetMember, err := h.memberRepo.GetByID(r.Context(), memberID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Member not found")
		return
	}

	if targetMember.Role == "owner" {
		respondError(w, http.StatusBadRequest, "Cannot remove the owner")
		return
	}

	if err := h.memberRepo.Delete(r.Context(), memberID); err != nil {
		log.Error().Err(err).Msg("Failed to remove member")
		respondError(w, http.StatusInternalServerError, "Failed to remove member")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Member removed successfully"})
}

// Helper functions

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = regexp.MustCompile(`[^a-z0-9-]+`).ReplaceAllString(slug, "-")
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
