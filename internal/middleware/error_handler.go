package middleware

import (
	"github.com/glennprays/golang-clean-arch-starter/internal/httperror"
	"github.com/glennprays/log"
	"github.com/gofiber/fiber/v2"
)

// ErrorHandler is the global Fiber error handler. It accepts the logger
// so unmatched or 5xx errors are recorded with the request's trace_id
// before the response is written — without that, production 500s would
// reach the client with no trace in the logs of what actually failed.
func ErrorHandler(logger *log.Logger) fiber.ErrorHandler {
	l := logger.With(log.String("component", "error_handler"))

	return func(c *fiber.Ctx, err error) error {
		traceID := GetTraceID(c)

		// 1. Fiber's own errors (404, 405, etc.) pass through untouched.
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{
				"error": fe.Message,
			})
		}

		// 2. Domain/application errors mapped to APIError.
		apiError := httperror.FromError(err)

		if apiError.Status == 0 {
			// Not a domain error — never mapped. Capture everything we
			// know about it before falling back to a generic 500.
			l.Error(traceID, "unhandled error reached global handler", map[string]any{
				"error":  err.Error(),
				"method": c.Method(),
				"path":   c.Path(),
			})
			apiError.Status = fiber.StatusInternalServerError
			apiError.Message = "Internal Server Error"
		} else if apiError.Status >= fiber.StatusInternalServerError {
			// Mapped 5xx — client gets a sanitized message, but we still
			// want the cause in the logs.
			l.Error(traceID, "internal domain error", map[string]any{
				"error":  err.Error(),
				"status": apiError.Status,
				"method": c.Method(),
				"path":   c.Path(),
			})
		}

		return c.Status(apiError.Status).JSON(fiber.Map{
			"error": apiError.Message,
		})
	}
}
