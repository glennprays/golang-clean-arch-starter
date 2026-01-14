package middleware

import (
	"log"

	"github.com/glennprays/golang-clean-arch-starter/internal/httperror"
	"github.com/gofiber/fiber/v2"
)

// ErrorHandler is a global error handler for Fiber
func ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		requestID := GetRequestID(c)

		// 1. Handle Fiber errors FIRST
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{
				"error":      fe.Message,
				"request_id": requestID,
			})
		}

		// 2. Handle domain/application errors
		apiError := httperror.FromError(err)

		if apiError.Status == 0 {
			apiError.Status = fiber.StatusInternalServerError
			apiError.Message = "Internal Server Error"
		}

		log.Printf("[RequestID: %s] Error: %v", requestID, err)

		return c.Status(apiError.Status).JSON(fiber.Map{
			"error":      apiError.Message,
			"request_id": requestID,
		})
	}
}
