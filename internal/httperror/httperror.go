// Package httperror defines the JSON contract for error responses.
//
// All error responses share the same shape regardless of source:
//
//	{
//	  "error": {
//	    "code":     "NOT_FOUND",
//	    "message":  "user 42 not found",
//	    "trace_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
//	    "details":  [ { "field": "...", "rule": "...", "message": "..." } ]
//	  }
//	}
//
// `code` is machine-readable (apperror.Kind). `message` is safe to show
// users. `trace_id` lets support correlate a report with server logs.
// `details` is omitted when empty and used for field-level validation.
package httperror

import "github.com/glennprays/golang-clean-arch-starter/internal/apperror"

// ErrorResponse is the top-level wrapper. The single-key envelope makes
// it easy to distinguish from any success body.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody describes one error.
type ErrorBody struct {
	Code    apperror.Kind         `json:"code"`
	Message string                `json:"message"`
	TraceID string                `json:"trace_id"`
	Details []apperror.FieldError `json:"details,omitempty"`
}
