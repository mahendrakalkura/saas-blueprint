package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mahendrakalkura/saas-blueprint/internal/auth"
	"github.com/mahendrakalkura/saas-blueprint/internal/search"
)

type SearchHandler struct {
	searchService *search.Service
}

func NewSearchHandler(searchService *search.Service) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

// Search godoc
// @Summary Global search
// @Description Search across all entities (users, organizations, notifications)
// @Tags search
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Results per page" default(20)
// @Param entity_types query string false "Comma-separated entity types to filter (user,organization,notification)"
// @Param from_date query string false "Filter results from date (RFC3339)"
// @Param to_date query string false "Filter results to date (RFC3339)"
// @Success 200 {object} search.SearchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /search [get]
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	// Get user context
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get query parameter
	query := r.URL.Query().Get("q")
	if query == "" {
		respondError(w, http.StatusBadRequest, "Search query is required")
		return
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get entity types filter
	var entityTypes []string
	if entityTypesParam := r.URL.Query().Get("entity_types"); entityTypesParam != "" {
		entityTypes = strings.Split(entityTypesParam, ",")
		for i := range entityTypes {
			entityTypes[i] = strings.TrimSpace(entityTypes[i])
		}
	}

	// Get date filters
	var fromDate, toDate time.Time
	if fromDateParam := r.URL.Query().Get("from_date"); fromDateParam != "" {
		if parsed, err := time.Parse(time.RFC3339, fromDateParam); err == nil {
			fromDate = parsed
		}
	}
	if toDateParam := r.URL.Query().Get("to_date"); toDateParam != "" {
		if parsed, err := time.Parse(time.RFC3339, toDateParam); err == nil {
			toDate = parsed
		}
	}

	// Build filters
	filters := search.SearchFilters{
		EntityTypes: entityTypes,
		FromDate:    fromDate,
		ToDate:      toDate,
		Page:        page,
		PageSize:    pageSize,
		UserOrgID:   &userCtx.OrganizationID,
	}

	// Perform search
	results, err := h.searchService.Search(r.Context(), query, filters)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to perform search")
		return
	}

	respondJSON(w, http.StatusOK, results)
}

// SearchUsers godoc
// @Summary Search users
// @Description Search specifically for users
// @Tags search
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Maximum results" default(20)
// @Success 200 {array} search.SearchResult
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /search/users [get]
func (h *SearchHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	// Get user context
	_, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get query parameter
	query := r.URL.Query().Get("q")
	if query == "" {
		respondError(w, http.StatusBadRequest, "Search query is required")
		return
	}

	// Get limit parameter
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Perform search
	results, err := h.searchService.SearchUsers(r.Context(), query, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to search users")
		return
	}

	respondJSON(w, http.StatusOK, results)
}

// SearchOrganizations godoc
// @Summary Search organizations
// @Description Search specifically for organizations the user has access to
// @Tags search
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Maximum results" default(20)
// @Success 200 {array} search.SearchResult
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /search/organizations [get]
func (h *SearchHandler) SearchOrganizations(w http.ResponseWriter, r *http.Request) {
	// Get user context
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get query parameter
	query := r.URL.Query().Get("q")
	if query == "" {
		respondError(w, http.StatusBadRequest, "Search query is required")
		return
	}

	// Get limit parameter
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Perform search
	results, err := h.searchService.SearchOrganizations(r.Context(), query, &userCtx.UserID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to search organizations")
		return
	}

	respondJSON(w, http.StatusOK, results)
}

// SearchNotifications godoc
// @Summary Search notifications
// @Description Search specifically for the user's notifications
// @Tags search
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Maximum results" default(20)
// @Success 200 {array} search.SearchResult
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /search/notifications [get]
func (h *SearchHandler) SearchNotifications(w http.ResponseWriter, r *http.Request) {
	// Get user context
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get query parameter
	query := r.URL.Query().Get("q")
	if query == "" {
		respondError(w, http.StatusBadRequest, "Search query is required")
		return
	}

	// Get limit parameter
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Perform search
	results, err := h.searchService.SearchNotifications(r.Context(), query, userCtx.UserID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to search notifications")
		return
	}

	respondJSON(w, http.StatusOK, results)
}

// GetSearchStatistics godoc
// @Summary Get search statistics
// @Description Get statistics about indexed content
// @Tags search
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /search/statistics [get]
func (h *SearchHandler) GetSearchStatistics(w http.ResponseWriter, r *http.Request) {
	// Get user context
	_, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get statistics
	stats, err := h.searchService.GetSearchStatistics(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get search statistics")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// RebuildSearchIndex godoc
// @Summary Rebuild search index
// @Description Rebuild the search index for a specific table (admin only)
// @Tags search
// @Accept json
// @Produce json
// @Param table query string true "Table name (users, organizations, notifications)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /search/rebuild [post]
func (h *SearchHandler) RebuildSearchIndex(w http.ResponseWriter, r *http.Request) {
	// Get user context
	userCtx, ok := auth.GetUserFromContext(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check if user is admin (you may want to implement proper admin check)
	// For now, we'll just allow any authenticated user
	// In production, add proper admin role check
	_ = userCtx

	// Get table parameter
	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		respondError(w, http.StatusBadRequest, "Table name is required")
		return
	}

	// Validate table name
	validTables := map[string]bool{
		"users":         true,
		"organizations": true,
		"notifications": true,
	}

	if !validTables[tableName] {
		respondError(w, http.StatusBadRequest, "Invalid table name")
		return
	}

	// Rebuild index
	err := h.searchService.RebuildSearchIndex(r.Context(), tableName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to rebuild search index")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Search index rebuilt successfully",
		"table":   tableName,
	})
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}

type ErrorResponse struct {
	Error string `json:"error"`
}
