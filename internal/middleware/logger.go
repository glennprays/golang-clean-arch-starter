package middleware

import (
	"time"

	"github.com/glennprays/log"
	"github.com/gofiber/fiber/v2"
)

// requestStartKey stores the time the request was first observed by
// NewHTTPLogger. ErrorHandler uses it (via requestLatency) so the
// error-path log line has the same latency field shape as the
// success-path one.
const requestStartKey = "_middleware_request_start"

// NewHTTPLogger emits one "http request" log per successful request.
// When c.Next() returns an error, the global ErrorHandler emits the
// log line instead — Fiber's ErrorHandler config runs after the
// middleware chain unwinds, so reading c.Response().StatusCode()
// here would still report the default 200 for routes that ended in
// a mapped error (e.g., 404 from a missing route).
func NewHTTPLogger(l *log.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(requestStartKey, time.Now())

		err := c.Next()

		if err != nil {
			// ErrorHandler will log this request with the actual
			// status it writes.
			return err
		}

		logHTTPRequest(l, c, GetTraceID(c), c.Response().StatusCode(), requestLatency(c))
		return nil
	}
}

// logHTTPRequest is the shared emitter for the per-request log line.
// Called from NewHTTPLogger (success) and ErrorHandler (error), so
// every request produces exactly one such line with consistent fields.
func logHTTPRequest(l *log.Logger, c *fiber.Ctx, traceID string, status int, latency time.Duration) {
	l.Info(traceID, "http request", map[string]any{
		"status":  status,
		"method":  c.Method(),
		"path":    c.Path(),
		"ip":      c.IP(),
		"latency": latency.String(),
	})
}

// requestLatency reads the request start time NewHTTPLogger stored in
// c.Locals and returns the elapsed duration. Returns 0 when the
// middleware isn't in the chain (e.g., from tests that exercise
// ErrorHandler directly).
func requestLatency(c *fiber.Ctx) time.Duration {
	if t, ok := c.Locals(requestStartKey).(time.Time); ok {
		return time.Since(t)
	}
	return 0
}
