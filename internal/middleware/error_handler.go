package middleware

import (
	"log"

	"github.com/glennprays/golang-clean-arch-starter/internal/httperror"
	"github.com/gofiber/fiber/v2"
)

// ErrorHandler is a global error handler for Fiber
func ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Get request ID for logging
		requestID := GetRequestID(c)

		// Convert domain errors to HTTP errors
		apiError := httperror.FromError(err)

		// Default to 500 if status not set
		if apiError.Status == 0 {
			apiError.Status = fiber.StatusInternalServerError
			apiError.Message = "Internal Server Error"
		}

		// Log error with request ID
		log.Printf("[RequestID: %s] Error: %v", requestID, err)

		// Return JSON error response
		return c.Status(apiError.Status).JSON(fiber.Map{
			"error":      apiError.Message,
			"request_id": requestID,
		})
	}
}
