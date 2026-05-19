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
// `Page` is populated only for paginated list responses (see Page docs).
type Envelope struct {
	Data    any        `json:"data,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
	Page    *Page      `json:"page,omitempty"`
	TraceID string     `json:"trace_id"`
}

// ErrorBody describes one error in the envelope.
type ErrorBody struct {
	Code    apperror.Kind         `json:"code"`
	Message string                `json:"message"`
	Details []apperror.FieldError `json:"details,omitempty"`
}

// Page carries pagination metadata for a list response. Only the
// fields relevant to the chosen scheme are populated.
//
//   - Cursor (Stripe-shaped): use NextCursor / PrevCursor.
//   - Offset (classic):       use Page / PageSize / Total.
//
// The Envelope's `page` field is omitted entirely when no Page is set.
type Page struct {
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`

	Page     int   `json:"page,omitempty"`
	PageSize int   `json:"page_size,omitempty"`
	Total    int64 `json:"total,omitempty"`
}

// OK writes a 200 success envelope wrapping the given payload. Use
// this from handlers so the response shape stays consistent.
func OK(c *fiber.Ctx, data any) error {
	return c.JSON(Envelope{
		Data:    data,
		TraceID: logctx.TraceID(c.UserContext()),
	})
}

// OKPaginated writes a 200 success envelope wrapping a list payload
// along with pagination metadata.
func OKPaginated(c *fiber.Ctx, data any, page Page) error {
	return c.JSON(Envelope{
		Data:    data,
		Page:    &page,
		TraceID: logctx.TraceID(c.UserContext()),
	})
}
