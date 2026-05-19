package middleware

import (
	"github.com/glennprays/golang-clean-arch-starter/internal/apperror"
	"github.com/glennprays/golang-clean-arch-starter/internal/httperror"
	"github.com/glennprays/log"
	"github.com/gofiber/fiber/v2"
)

// ErrorHandler is the global Fiber error handler. It accepts the logger
// so 5xx and unmatched errors are recorded with the request's trace_id
// before the response is written — without that, production 500s would
// reach the client with no trace in the logs of what actually failed.
func ErrorHandler(logger *log.Logger) fiber.ErrorHandler {
	l := logger.With(log.String("component", "error_handler"))

	return func(c *fiber.Ctx, err error) error {
		traceID := GetTraceID(c)

		// 1. Application errors (typed) — map by Kind.
		if ae := apperror.AsAppError(err); ae != nil {
			status := ae.HTTPStatus()
			if status >= fiber.StatusInternalServerError {
				cause := ""
				if ae.Cause != nil {
					cause = ae.Cause.Error()
				}
				l.Error(traceID, "internal app error", map[string]any{
					"kind":   string(ae.Kind),
					"msg":    ae.Message,
					"cause":  cause,
					"method": c.Method(),
					"path":   c.Path(),
				})
			}
			return writeError(c, status, httperror.ErrorBody{
				Code:    ae.Kind,
				Message: ae.Message,
				TraceID: traceID,
				Details: ae.Details,
			})
		}

		// 2. Fiber's own errors (404, 405, etc.) — pass status through,
		// translate the status to a Kind so the response shape stays
		// consistent.
		if fe, ok := err.(*fiber.Error); ok {
			return writeError(c, fe.Code, httperror.ErrorBody{
				Code:    fiberKind(fe.Code),
				Message: fe.Message,
				TraceID: traceID,
			})
		}

		// 3. Unknown error — log everything, expose nothing.
		l.Error(traceID, "unhandled error reached global handler", map[string]any{
			"error":  err.Error(),
			"method": c.Method(),
			"path":   c.Path(),
		})
		return writeError(c, fiber.StatusInternalServerError, httperror.ErrorBody{
			Code:    apperror.KindInternal,
			Message: "internal server error",
			TraceID: traceID,
		})
	}
}

func writeError(c *fiber.Ctx, status int, body httperror.ErrorBody) error {
	return c.Status(status).JSON(httperror.ErrorResponse{Error: body})
}

func fiberKind(code int) apperror.Kind {
	switch code {
	case fiber.StatusNotFound:
		return apperror.KindNotFound
	case fiber.StatusMethodNotAllowed, fiber.StatusBadRequest:
		return apperror.KindBadRequest
	case fiber.StatusUnauthorized:
		return apperror.KindUnauthorized
	case fiber.StatusForbidden:
		return apperror.KindForbidden
	case fiber.StatusConflict:
		return apperror.KindConflict
	default:
		return apperror.KindInternal
	}
}
