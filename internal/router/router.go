package router

import (
	"github.com/glennprays/golang-clean-arch-starter/internal/handler"
	"github.com/glennprays/golang-clean-arch-starter/internal/middleware"
	"github.com/glennprays/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Router struct {
	logger        *log.Logger
	HealthHandler *handler.HealthHandler
}

func NewRouter(
	logger *log.Logger,
	healthHandler *handler.HealthHandler,
) *Router {
	routerLogger := logger.With(log.String("component", "router"))
	return &Router{
		logger:        routerLogger,
		HealthHandler: healthHandler,
	}
}

// Setup configures all application routes
func (r *Router) Setup(app *fiber.App) {
	// Global middleware. TraceID must run before recover so that any
	// panic surfaced by the recovery middleware can be correlated by
	// trace_id in the logs.
	app.Use(middleware.TraceID())
	app.Use(recover.New())
	app.Use(middleware.CORS())

	app.Use(middleware.NewHTTPLogger(r.logger))

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
