package router

import (
	"github.com/glennprays/golang-clean-arch-starter/internal/handler"
	"github.com/glennprays/golang-clean-arch-starter/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

type Router struct {
	HealthHandler *handler.HealthHandler
}

func NewRouter(healthHandler *handler.HealthHandler) *Router {
	return &Router{
		HealthHandler: healthHandler,
	}
}

// Setup configures all application routes
func (r *Router) Setup(app *fiber.App) {
	// Global middleware
	app.Use(middleware.RequestID())
	app.Use(middleware.CORS())

	// API v1 group
	v1 := app.Group("/api/v1")

	// Health routes
	r.setupHealthRoutes(v1)

	// Future route groups can be added here:
	// r.setupUserRoutes(v1)
	// r.setupAuthRoutes(v1)
}

func (r *Router) setupHealthRoutes(group fiber.Router) {
	group.Get("/health", r.HealthHandler.Check)
}
