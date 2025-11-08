package middleware

import (
	"context"
	"net/http"
	"strconv"
)

type PaginationParams struct {
	Page     int
	PageSize int
	Offset   int
}

type paginationKey struct{}

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Pagination middleware extracts pagination params from query string
func Pagination(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := getIntParam(r, "page", DefaultPage)
		if page < 1 {
			page = DefaultPage
		}

		pageSize := getIntParam(r, "page_size", DefaultPageSize)
		if pageSize < 1 {
			pageSize = DefaultPageSize
		}
		if pageSize > MaxPageSize {
			pageSize = MaxPageSize
		}

		offset := (page - 1) * pageSize

		params := PaginationParams{
			Page:     page,
			PageSize: pageSize,
			Offset:   offset,
		}

		ctx := context.WithValue(r.Context(), paginationKey{}, params)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetPagination retrieves pagination params from context
func GetPagination(r *http.Request) PaginationParams {
	params, ok := r.Context().Value(paginationKey{}).(PaginationParams)
	if !ok {
		return PaginationParams{
			Page:     DefaultPage,
			PageSize: DefaultPageSize,
			Offset:   0,
		}
	}
	return params
}

func getIntParam(r *http.Request, key string, defaultValue int) int {
	valueStr := r.URL.Query().Get(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// PaginatedResponse is a standard paginated response structure
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalItems int64       `json:"total_items"`
	TotalPages int         `json:"total_pages"`
}

func NewPaginatedResponse(data interface{}, page, pageSize int, totalItems int64) PaginatedResponse {
	totalPages := int(totalItems) / pageSize
	if int(totalItems)%pageSize != 0 {
		totalPages++
	}

	return PaginatedResponse{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}
