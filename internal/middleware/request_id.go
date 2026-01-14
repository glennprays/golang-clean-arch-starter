package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	RequestIDHeader     = "X-Request-ID"
	RequestIDContextKey = "request_id"
)

// RequestID middleware handles request ID generation and propagation
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Try to read from incoming request header
		requestID := c.Get(RequestIDHeader)

		// If not present, generate new UUID
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store in context for use throughout request lifecycle
		c.Locals(RequestIDContextKey, requestID)

		// Add to response headers
		c.Set(RequestIDHeader, requestID)

		return c.Next()
	}
}

// GetRequestID retrieves request ID from context
func GetRequestID(c *fiber.Ctx) string {
	if id, ok := c.Locals(RequestIDContextKey).(string); ok {
		return id
	}
	return ""
}
