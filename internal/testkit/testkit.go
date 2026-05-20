// Package testkit provides test helpers for spinning up a Fiber app
// with the same middleware stack the production server uses.
//
// Use it from handler tests so the response envelope, trace ID, and
// error handler are exercised end-to-end without duplicating wiring:
//
//	func TestHealthHandler(t *testing.T) {
//	    h := handler.NewHealthHandler()
//	    app := testkit.NewTestApp(t, func(app *fiber.App) {
//	        app.Get("/health", h.Check)
//	    })
//	    resp, _ := app.Test(httptest.NewRequest(http.MethodGet, "/health", nil))
//	    // assert status, headers, JSON envelope...
//	}
package testkit

import (
	"testing"

	"github.com/glennprays/golang-clean-arch-starter/internal/middleware"
	"github.com/glennprays/log"
	"github.com/gofiber/fiber/v2"
)

// NewTestApp returns a Fiber app wired with the production middleware
// stack (TraceID, ErrorHandler) and a quiet logger filtered to errors
// only. The register callback mounts whatever routes the test needs.
func NewTestApp(t *testing.T, register func(*fiber.App)) *fiber.App {
	t.Helper()

	logger, err := log.New(log.Config{
		Service: "test",
		Env:     "development",
		Level:   log.Level("error"),
		Output:  log.OutputType("stdout"),
	})
	if err != nil {
		t.Fatalf("create test logger: %v", err)
	}

	app := fiber.New(fiber.Config{
		ErrorHandler:          middleware.ErrorHandler(logger),
		DisableStartupMessage: true,
	})
	app.Use(middleware.TraceID())

	register(app)
	return app
}
