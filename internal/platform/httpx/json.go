// Package httpx provides minimal HTTP helpers: JSON I/O, error responses,
// validation helpers, and the Middleware contract.
package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const maxBodyBytes = 1 << 20 // 1 MiB

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

func ReadJSON(r *http.Request, dst any) error {
	limited := io.LimitReader(r.Body, maxBodyBytes)
	return json.NewDecoder(limited).Decode(dst)
}

// Generic sentinel errors. Wrap with fmt.Errorf("...: %w", ErrX) and
// dispatch via WriteServiceError. Domain-specific errors stay in their
// own handler.go.
var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
	ErrValidation   = errors.New("validation failed")
)

// WriteServiceError maps wrapped sentinels to status codes; non-sentinel
// errors return 500 with a generic message.
func WriteServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrUnauthorized):
		WriteError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, ErrForbidden):
		WriteError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrConflict):
		WriteError(w, http.StatusConflict, err.Error())
	case errors.Is(err, ErrValidation):
		WriteError(w, http.StatusUnprocessableEntity, err.Error())
	default:
		WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
