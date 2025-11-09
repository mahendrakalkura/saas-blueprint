package search

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/mahendrakalkura/saas-blueprint/internal/database"
)

// SearchResult represents a single search result
type SearchResult struct {
	EntityType  string    `json:"entity_type"`
	EntityID    string    `json:"entity_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Rank        float64   `json:"rank"`
	CreatedAt   time.Time `json:"created_at"`
	Highlight   string    `json:"highlight,omitempty"`
}

// SearchResponse represents the complete search response
type SearchResponse struct {
	Query      string         `json:"query"`
	Results    []SearchResult `json:"results"`
	TotalCount int            `json:"total_count"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	HasMore    bool           `json:"has_more"`
}

// SearchFilters represents filters for search
type SearchFilters struct {
	EntityTypes []string  // Filter by entity types (user, organization, notification)
	FromDate    time.Time // Filter results from this date
	ToDate      time.Time // Filter results up to this date
	Page        int       // Page number (1-indexed)
	PageSize    int       // Results per page
	UserOrgID   *string   // Organization ID for permission filtering
}

// Service provides search functionality
type Service struct {
	db *database.DB
}

// NewService creates a new search service
func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

// Search performs a global search across all searchable entities
func (s *Service) Search(ctx context.Context, query string, filters SearchFilters) (*SearchResponse, error) {
	// Set defaults
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 || filters.PageSize > 100 {
		filters.PageSize = 20
	}

	// Clean and validate query
	query = strings.TrimSpace(query)
	if query == "" {
		return &SearchResponse{
			Query:      query,
			Results:    []SearchResult{},
			TotalCount: 0,
			Page:       filters.Page,
			PageSize:   filters.PageSize,
			HasMore:    false,
		}, nil
	}

	// Build the search query
	offset := (filters.Page - 1) * filters.PageSize
	limit := filters.PageSize + 1 // Fetch one extra to check if there are more results

	sqlQuery := `
		SELECT
			entity_type,
			entity_id,
			title,
			description,
			rank,
			created_at
		FROM global_search($1, $2, $3)
	`

	var args []interface{}
	args = append(args, query)

	// Add user organization ID filter
	if filters.UserOrgID != nil {
		args = append(args, *filters.UserOrgID)
	} else {
		args = append(args, nil)
	}

	args = append(args, limit)

	// Execute the search
	rows, err := s.db.DB.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		err := rows.Scan(
			&result.EntityType,
			&result.EntityID,
			&result.Title,
			&result.Description,
			&result.Rank,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}

		// Apply filters
		if len(filters.EntityTypes) > 0 {
			found := false
			for _, entityType := range filters.EntityTypes {
				if result.EntityType == entityType {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		if !filters.FromDate.IsZero() && result.CreatedAt.Before(filters.FromDate) {
			continue
		}

		if !filters.ToDate.IsZero() && result.CreatedAt.After(filters.ToDate) {
			continue
		}

		// Generate highlight snippet
		result.Highlight = s.generateHighlight(result.Description, query)

		results = append(results, result)

		// Stop if we have enough results
		if len(results) >= limit {
			break
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating search results: %w", err)
	}

	// Check if there are more results
	hasMore := len(results) > filters.PageSize
	if hasMore {
		results = results[:filters.PageSize]
	}

	// Get total count (approximate for performance)
	totalCount := (filters.Page-1)*filters.PageSize + len(results)
	if hasMore {
		totalCount++
	}

	return &SearchResponse{
		Query:      query,
		Results:    results,
		TotalCount: totalCount,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		HasMore:    hasMore,
	}, nil
}

// SearchUsers performs a search specifically for users
func (s *Service) SearchUsers(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}

	sqlQuery := `
		SELECT
			id,
			name,
			email,
			ts_rank(search_vector, websearch_to_tsquery('english_unaccent', $1)) as rank,
			created_at
		FROM users
		WHERE search_vector @@ websearch_to_tsquery('english_unaccent', $1)
		ORDER BY rank DESC, created_at DESC
		LIMIT $2
	`

	rows, err := s.db.DB.QueryContext(ctx, sqlQuery, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		result.EntityType = "user"

		err := rows.Scan(
			&result.EntityID,
			&result.Title,
			&result.Description,
			&result.Rank,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user result: %w", err)
		}

		result.Highlight = s.generateHighlight(result.Description, query)
		results = append(results, result)
	}

	return results, rows.Err()
}

// SearchOrganizations performs a search specifically for organizations
func (s *Service) SearchOrganizations(ctx context.Context, query string, userID *string, limit int) ([]SearchResult, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// If userID is provided, only return organizations the user is a member of
	sqlQuery := `
		SELECT
			o.id,
			o.name,
			COALESCE(o.description, '') as description,
			ts_rank(o.search_vector, websearch_to_tsquery('english_unaccent', $1)) as rank,
			o.created_at
		FROM organizations o
	`

	var args []interface{}
	args = append(args, query)

	if userID != nil {
		sqlQuery += `
			INNER JOIN organization_members om ON o.id = om.organization_id
			WHERE o.search_vector @@ websearch_to_tsquery('english_unaccent', $1)
				AND om.user_id = $2
		`
		args = append(args, *userID)
		sqlQuery += " ORDER BY rank DESC, o.created_at DESC LIMIT $3"
		args = append(args, limit)
	} else {
		sqlQuery += `
			WHERE o.search_vector @@ websearch_to_tsquery('english_unaccent', $1)
			ORDER BY rank DESC, o.created_at DESC
			LIMIT $2
		`
		args = append(args, limit)
	}

	rows, err := s.db.DB.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search organizations: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		result.EntityType = "organization"

		err := rows.Scan(
			&result.EntityID,
			&result.Title,
			&result.Description,
			&result.Rank,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization result: %w", err)
		}

		result.Highlight = s.generateHighlight(result.Description, query)
		results = append(results, result)
	}

	return results, rows.Err()
}

// SearchNotifications performs a search specifically for notifications
func (s *Service) SearchNotifications(ctx context.Context, query string, userID string, limit int) ([]SearchResult, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}

	sqlQuery := `
		SELECT
			id,
			title,
			message,
			ts_rank(search_vector, websearch_to_tsquery('english_unaccent', $1)) as rank,
			created_at
		FROM notifications
		WHERE search_vector @@ websearch_to_tsquery('english_unaccent', $1)
			AND user_id = $2
		ORDER BY rank DESC, created_at DESC
		LIMIT $3
	`

	rows, err := s.db.DB.QueryContext(ctx, sqlQuery, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search notifications: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		result.EntityType = "notification"

		err := rows.Scan(
			&result.EntityID,
			&result.Title,
			&result.Description,
			&result.Rank,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification result: %w", err)
		}

		result.Highlight = s.generateHighlight(result.Description, query)
		results = append(results, result)
	}

	return results, rows.Err()
}

// GetSearchStatistics returns statistics about indexed content
func (s *Service) GetSearchStatistics(ctx context.Context) (map[string]interface{}, error) {
	sqlQuery := `SELECT table_name, total_records, indexed_records FROM search_statistics`

	rows, err := s.db.DB.QueryContext(ctx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get search statistics: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]interface{})
	for rows.Next() {
		var tableName string
		var totalRecords, indexedRecords int

		err := rows.Scan(&tableName, &totalRecords, &indexedRecords)
		if err != nil {
			return nil, fmt.Errorf("failed to scan statistics: %w", err)
		}

		stats[tableName] = map[string]int{
			"total":   totalRecords,
			"indexed": indexedRecords,
		}
	}

	return stats, rows.Err()
}

// generateHighlight creates a highlighted snippet of the text containing the search query
func (s *Service) generateHighlight(text, query string) string {
	if text == "" || query == "" {
		return ""
	}

	// Convert to lowercase for case-insensitive search
	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)

	// Find the position of the query
	pos := strings.Index(lowerText, lowerQuery)
	if pos == -1 {
		// If exact match not found, just return beginning of text
		if len(text) > 150 {
			return text[:150] + "..."
		}
		return text
	}

	// Calculate snippet boundaries
	start := pos - 50
	if start < 0 {
		start = 0
	}

	end := pos + len(query) + 50
	if end > len(text) {
		end = len(text)
	}

	// Extract snippet
	snippet := text[start:end]

	// Add ellipsis
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(text) {
		snippet = snippet + "..."
	}

	return snippet
}

// RebuildSearchIndex rebuilds the search index for a specific table
func (s *Service) RebuildSearchIndex(ctx context.Context, tableName string) error {
	var sqlQuery string

	switch tableName {
	case "users":
		sqlQuery = `
			UPDATE users SET search_vector =
				setweight(to_tsvector('english_unaccent', COALESCE(name, '')), 'A') ||
				setweight(to_tsvector('english_unaccent', COALESCE(email, '')), 'B')
		`
	case "organizations":
		sqlQuery = `
			UPDATE organizations SET search_vector =
				setweight(to_tsvector('english_unaccent', COALESCE(name, '')), 'A') ||
				setweight(to_tsvector('english_unaccent', COALESCE(description, '')), 'C')
		`
	case "notifications":
		sqlQuery = `
			UPDATE notifications SET search_vector =
				setweight(to_tsvector('english_unaccent', COALESCE(title, '')), 'A') ||
				setweight(to_tsvector('english_unaccent', COALESCE(message, '')), 'B')
		`
	default:
		return fmt.Errorf("unsupported table: %s", tableName)
	}

	_, err := s.db.DB.ExecContext(ctx, sqlQuery)
	if err != nil {
		return fmt.Errorf("failed to rebuild search index for %s: %w", tableName, err)
	}

	return nil
}
