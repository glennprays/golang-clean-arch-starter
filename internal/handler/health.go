package handler

import (
	"time"

	"github.com/glennprays/golang-clean-arch-starter/internal/httperror"
	"github.com/gofiber/fiber/v2"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

func (h *HealthHandler) Check(c *fiber.Ctx) error {
	return httperror.OK(c, HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}
