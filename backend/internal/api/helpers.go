// Package api provides HTTP handlers and route registration for the WMS API.
package api

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"

	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// PathUUID extracts a UUID path parameter from the request.
func PathUUID(r *http.Request, key string) (uuid.UUID, error) {
	raw := r.PathValue(key)
	if raw == "" {
		return uuid.Nil, pkgerrors.NewInvalidInput("missing path parameter: " + key)
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, pkgerrors.NewInvalidInput("invalid UUID format for " + key)
	}
	return id, nil
}

// QueryParam returns a query string parameter value or the default.
func QueryParam(r *http.Request, key, defaultVal string) string {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}
	return v
}

// QueryParamInt returns a query string parameter as an int or the default.
func QueryParamInt(r *http.Request, key string, defaultVal int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return defaultVal
	}
	return n
}

// PaginationParams extracts page and page_size from query parameters with defaults.
// page defaults to 1, page_size defaults to 20, max page_size is 100.
func PaginationParams(r *http.Request) (page, pageSize int) {
	page = QueryParamInt(r, "page", 1)
	if page < 1 {
		page = 1
	}

	pageSize = QueryParamInt(r, "page_size", 20)
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return page, pageSize
}

// paginationOffset computes the SQL offset from page and pageSize (0-indexed).
func paginationOffset(page, pageSize int) int {
	return (page - 1) * pageSize
}
