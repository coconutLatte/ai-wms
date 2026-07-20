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
