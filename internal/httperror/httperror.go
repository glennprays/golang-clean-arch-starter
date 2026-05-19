// Package httperror defines the JSON contract for HTTP responses.
//
// Every response (success or error) uses the same Envelope:
//
//	// Success
//	{ "data": { ... }, "trace_id": "f47ac10b-..." }
//
//	// Error
//	{ "error": { "code": "NOT_FOUND", "message": "...", "details": [...] },
//	  "trace_id": "f47ac10b-..." }
//
// Exactly one of `data` and `error` is populated; `trace_id` is always
// present so support can correlate a response with server logs.
//
// The package is named httperror for historical reasons — it owns
// both the success and error shapes now.
package httperror

import (
	"github.com/glennprays/golang-clean-arch-starter/internal/apperror"
	"github.com/glennprays/golang-clean-arch-starter/pkg/logctx"
	"github.com/gofiber/fiber/v2"
)

// Envelope is the top-level JSON shape returned by every handler.
type Envelope struct {
	Data    any        `json:"data,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
	TraceID string     `json:"trace_id"`
}

// ErrorBody describes one error in the envelope.
type ErrorBody struct {
	Code    apperror.Kind         `json:"code"`
	Message string                `json:"message"`
	Details []apperror.FieldError `json:"details,omitempty"`
}

// OK writes a 200 success envelope wrapping the given payload. Use
// this from handlers so the response shape stays consistent.
func OK(c *fiber.Ctx, data any) error {
	return c.JSON(Envelope{
		Data:    data,
		TraceID: logctx.TraceID(c.UserContext()),
	})
}
