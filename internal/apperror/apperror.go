// Package apperror is the application's typed error model.
//
// One Error type carries a Kind (machine-readable category), a
// human-readable message, an optional Cause (for logs, never sent
// to clients) and optional structured Details (validation, etc.).
//
// The design mirrors patterns used by k8s.io/apimachinery/pkg/api/errors,
// gocloud.dev/gcerrors, and google.golang.org/grpc/status — a single
// Kind enum with per-kind constructors. Use errors.Is for matching
// (the custom Is method compares by Kind, not by pointer identity).
package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// Kind is a stable, machine-readable error category. Each Kind maps
// 1:1 to one HTTP status. Values are stable strings so they can be
// returned to clients in the response `code` field.
type Kind string

const (
	KindBadRequest   Kind = "BAD_REQUEST"
	KindUnauthorized Kind = "UNAUTHORIZED"
	KindForbidden    Kind = "FORBIDDEN"
	KindNotFound     Kind = "NOT_FOUND"
	KindConflict     Kind = "CONFLICT"
	KindValidation   Kind = "VALIDATION_FAILED"
	KindInternal     Kind = "INTERNAL"
)

// FieldError describes one failed rule on one field. Used inside
// Error.Details for cases like request body validation.
type FieldError struct {
	Field   string `json:"field"`
	Rule    string `json:"rule,omitempty"`
	Message string `json:"message"`
}

// Error is the application-level error type. All methods are nil-safe.
type Error struct {
	Kind    Kind
	Message string
	Cause   error
	Details []FieldError
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Kind, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

// Unwrap returns the wrapped cause so the stdlib errors.Is/errors.As
// chain walkers can see through to it.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Is matches by Kind. This is what makes errors.Is(err, ErrNotFound)
// work even when err is a constructed *Error rather than the sentinel
// itself — both carry Kind=KindNotFound, so they compare equal.
func (e *Error) Is(target error) bool {
	if e == nil {
		return target == nil
	}
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Kind == t.Kind
}

// Wrap attaches a cause (for logs) and returns the same *Error so
// constructors chain: apperror.Internal("db failed").Wrap(err).
func (e *Error) Wrap(cause error) *Error {
	if e == nil {
		return nil
	}
	e.Cause = cause
	return e
}

// WithDetails attaches field-level errors (validation, etc.).
func (e *Error) WithDetails(d ...FieldError) *Error {
	if e == nil {
		return nil
	}
	e.Details = append(e.Details, d...)
	return e
}

// HTTPStatus maps Kind to the HTTP status code. Unknown kinds fall
// back to 500.
func (e *Error) HTTPStatus() int {
	if e == nil {
		return http.StatusInternalServerError
	}
	switch e.Kind {
	case KindBadRequest, KindValidation:
		return http.StatusBadRequest
	case KindUnauthorized:
		return http.StatusUnauthorized
	case KindForbidden:
		return http.StatusForbidden
	case KindNotFound:
		return http.StatusNotFound
	case KindConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// Sentinels are used as targets for errors.Is. Don't return these
// directly — they have no contextual message. Use the constructors
// below instead.
var (
	ErrBadRequest   = &Error{Kind: KindBadRequest, Message: "bad request"}
	ErrUnauthorized = &Error{Kind: KindUnauthorized, Message: "unauthorized"}
	ErrForbidden    = &Error{Kind: KindForbidden, Message: "forbidden"}
	ErrNotFound     = &Error{Kind: KindNotFound, Message: "not found"}
	ErrConflict     = &Error{Kind: KindConflict, Message: "conflict"}
	ErrValidation   = &Error{Kind: KindValidation, Message: "validation failed"}
	ErrInternal     = &Error{Kind: KindInternal, Message: "internal error"}
)

// Per-kind constructors. One line at the call site.

func BadRequest(msg string) *Error   { return &Error{Kind: KindBadRequest, Message: msg} }
func Unauthorized(msg string) *Error { return &Error{Kind: KindUnauthorized, Message: msg} }
func Forbidden(msg string) *Error    { return &Error{Kind: KindForbidden, Message: msg} }
func NotFound(msg string) *Error     { return &Error{Kind: KindNotFound, Message: msg} }
func Conflict(msg string) *Error     { return &Error{Kind: KindConflict, Message: msg} }
func Validation(msg string) *Error   { return &Error{Kind: KindValidation, Message: msg} }
func Internal(msg string) *Error     { return &Error{Kind: KindInternal, Message: msg} }

// Formatted variants. Save fmt.Sprintf at the call site.

func BadRequestf(format string, a ...any) *Error {
	return BadRequest(fmt.Sprintf(format, a...))
}
func Unauthorizedf(format string, a ...any) *Error {
	return Unauthorized(fmt.Sprintf(format, a...))
}
func Forbiddenf(format string, a ...any) *Error {
	return Forbidden(fmt.Sprintf(format, a...))
}
func NotFoundf(format string, a ...any) *Error {
	return NotFound(fmt.Sprintf(format, a...))
}
func Conflictf(format string, a ...any) *Error {
	return Conflict(fmt.Sprintf(format, a...))
}
func Validationf(format string, a ...any) *Error {
	return Validation(fmt.Sprintf(format, a...))
}
func Internalf(format string, a ...any) *Error {
	return Internal(fmt.Sprintf(format, a...))
}

// AsAppError extracts a *Error from any error chain. Returns nil if
// the chain doesn't contain one.
func AsAppError(err error) *Error {
	var ae *Error
	if errors.As(err, &ae) {
		return ae
	}
	return nil
}
