package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

// Error codes — sentinel values for machine-readable comparison.
const (
	ErrCodeValidation   = "VALIDATION_ERROR"
	ErrCodeUnauthorized = "UNAUTHORIZED"
	ErrCodeForbidden    = "FORBIDDEN"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeConflict     = "CONFLICT"
	ErrCodeRateLimited  = "RATE_LIMITED"
	ErrCodeInternal     = "INTERNAL_ERROR"
	ErrCodeBadRequest   = "BAD_REQUEST"
)

// AppError is the canonical error type for all application-layer errors.
type AppError struct {
	Code       string // machine-readable code (see constants above)
	Message    string // user-facing message in Portuguese
	HTTPStatus int    // HTTP status code
	Err        error  // wrapped original error (optional, for inspection)
}

func (e *AppError) Error() string { return e.Message }

func (e *AppError) Unwrap() error { return e.Err }

// HTTP status map for errors without explicit status.
var codeStatus = map[string]int{
	ErrCodeValidation:   http.StatusBadRequest,
	ErrCodeUnauthorized: http.StatusUnauthorized,
	ErrCodeForbidden:    http.StatusForbidden,
	ErrCodeNotFound:     http.StatusNotFound,
	ErrCodeConflict:     http.StatusConflict,
	ErrCodeRateLimited:  http.StatusTooManyRequests,
	ErrCodeInternal:     http.StatusInternalServerError,
	ErrCodeBadRequest:   http.StatusBadRequest,
}

// handleError dispatches an error to the appropriate response format.
// It always logs the error with structured fields before responding.
func (a *App) handleError(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		appErr = &AppError{
			Code:       ErrCodeInternal,
			Message:    "Erro interno do servidor",
			HTTPStatus: http.StatusInternalServerError,
			Err:        err,
		}
	}

	// Fill in default HTTP status from code if not set.
	if appErr.HTTPStatus == 0 {
		if s, ok := codeStatus[appErr.Code]; ok {
			appErr.HTTPStatus = s
		} else {
			appErr.HTTPStatus = http.StatusInternalServerError
		}
	}

	// Structured logging with context and fields.
	slog.ErrorContext(r.Context(), appErr.Message,
		"code", appErr.Code,
		"status", appErr.HTTPStatus,
		"path", r.URL.Path,
		"method", r.Method,
		"error", appErr.Err,
	)

	// HTMX request — return HTML regardless of path.
	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(appErr.HTTPStatus)
		err := a.tpl.ExecuteTemplate(w, "error_toast", map[string]any{
			"Message": appErr.Message,
			"Code":    appErr.Code,
		})
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to render error toast",
				"template_error", err,
			)
			http.Error(w, appErr.Message, appErr.HTTPStatus)
		}
		return
	}

	// API request — JSON error response.
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeJSON(w, appErr.HTTPStatus, map[string]string{
			"error": appErr.Message,
			"code":  appErr.Code,
		})
		return
	}

	// Regular page — simple text error.
	http.Error(w, appErr.Message, appErr.HTTPStatus)
}

// ctxRequestID is the context key for the request ID.
const ctxRequestID ctxKey = "request_id"

// requestIDMiddleware injects a request ID into the context and response header.
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			b := make([]byte, 8)
			rand.Read(b)
			id = fmt.Sprintf("%x", b)
		}
		ctx := context.WithValue(r.Context(), ctxRequestID, id)
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getRequestID retrieves the request ID from context.
func getRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(ctxRequestID).(string); ok {
		return id
	}
	return ""
}
