package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mahendrakalkura/saas-blueprint/internal/auth"
	"github.com/mahendrakalkura/saas-blueprint/internal/repository"
	"github.com/rs/zerolog/log"
)

type OrganizationContext struct {
	OrganizationID string
	MemberID       string
	Role           string
}

type contextKey string

const OrganizationContextKey contextKey = "organization"

// RequireOrganizationMember checks if user is a member of the organization in the URL
func RequireOrganizationMember(memberRepo *repository.OrganizationMemberRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx, ok := auth.GetUserFromContext(r)
			if !ok {
				http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			orgID := chi.URLParam(r, "id")
			if orgID == "" {
				http.Error(w, `{"error":"Organization ID required"}`, http.StatusBadRequest)
				return
			}

			member, err := memberRepo.GetByOrganizationAndUser(r.Context(), orgID, userCtx.UserID)
			if err != nil {
				http.Error(w, `{"error":"Not a member of this organization"}`, http.StatusForbidden)
				return
			}

			// Add organization context to request
			ctx := context.WithValue(r.Context(), OrganizationContextKey, &OrganizationContext{
				OrganizationID: orgID,
				MemberID:       member.ID,
				Role:           member.Role,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireOrganizationRole checks if user has a specific role or higher
func RequireOrganizationRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgCtx := GetOrganizationFromContext(r)
			if orgCtx == nil {
				log.Error().Msg("Organization context not found - ensure RequireOrganizationMember runs first")
				http.Error(w, `{"error":"Organization context required"}`, http.StatusInternalServerError)
				return
			}

			// Check if user has one of the allowed roles
			hasRole := false
			for _, role := range allowedRoles {
				if orgCtx.Role == role {
					hasRole = true
					break
				}
			}

			// Owner always has access
			if orgCtx.Role == "owner" {
				hasRole = true
			}

			if !hasRole {
				http.Error(w, `{"error":"Insufficient permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireOwner checks if user is the organization owner
func RequireOwner() func(http.Handler) http.Handler {
	return RequireOrganizationRole("owner")
}

// RequireAdmin checks if user is at least an admin
func RequireAdmin() func(http.Handler) http.Handler {
	return RequireOrganizationRole("owner", "admin")
}

// GetOrganizationFromContext retrieves organization context from request
func GetOrganizationFromContext(r *http.Request) *OrganizationContext {
	orgCtx, ok := r.Context().Value(OrganizationContextKey).(*OrganizationContext)
	if !ok {
		return nil
	}
	return orgCtx
}
