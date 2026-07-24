// Package api provides HTTP handlers and route registration for the WMS API.
package api

import (
	"encoding/json"
	"math"
	"net/http"

	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"
)

// PaginationMeta holds pagination metadata for list endpoints.
type PaginationMeta struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalPages int `json:"total_pages"`
}

// NewPaginationMeta computes pagination metadata from the total count, page, and page size.
func NewPaginationMeta(total, page, pageSize int) PaginationMeta {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}
	return PaginationMeta{
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

// ListResponse wraps a paginated list of items with pagination metadata.
type ListResponse[T any] struct {
	Data       []T            `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// WriteJSON writes a JSON response with the given status code and body.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// WriteCreated writes a 201 Created JSON response.
func WriteCreated(w http.ResponseWriter, data any) {
	WriteJSON(w, http.StatusCreated, data)
}

// WriteNoContent writes a 204 No Content response.
func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// WriteError writes a standardized RFC 7807 Problem Details error response.
// It maps the error to the correct HTTP status code and formats the response
// using problem details. This is THE single error-writing function that all
// handlers should use.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	pd := pkgerrors.NewProblemDetail(err, r.URL.Path)
	if pd == nil {
		// Fallback: if NewProblemDetail returns nil for a nil error, write 500.
		WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
		return
	}

	WriteJSON(w, pd.Status, pd)
}

// ReadJSON reads and decodes a JSON request body into the provided value.
// Returns an error with context if decoding fails.
func ReadJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return pkgerrors.NewInvalidInput("invalid request body: " + err.Error())
	}
	return nil
}
