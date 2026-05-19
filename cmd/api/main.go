package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/glennprays/golang-clean-arch-starter/config"
	"github.com/glennprays/golang-clean-arch-starter/internal/infrastructure"
	"github.com/glennprays/golang-clean-arch-starter/internal/middleware"
	"github.com/glennprays/log"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func main() {
	lifecycleID := uuid.New().String()
	// Initialize app dependencies via Wire
	app, err := infrastructure.InitializeApp()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize app: %v", err))
	}
	defer app.Logger.Sync()

	logger := app.Logger.With(log.String("component", "main"))

	// Create Fiber app with custom error handler and explicit limits.
	// Defaults (15s/15s read/write, 4 MiB body) are not appropriate for
	// production — slow clients can tie up connections and large bodies
	// reach handlers without backpressure.
	fiberApp := fiber.New(fiber.Config{
		AppName:               app.Config.AppName,
		ErrorHandler:          middleware.ErrorHandler(app.Logger),
		DisableStartupMessage: true,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		IdleTimeout:           120 * time.Second,
		BodyLimit:             1 << 20, // 1 MiB
	})

	// Setup routes (includes global middleware in correct order)
	app.Router.Setup(fiberApp)

	// Start server in goroutine
	addr := fmt.Sprintf(":%d", app.Config.AppPort)
	go func() {
		logger.Info(lifecycleID, "Starting server", map[string]any{
			"address":  addr,
			"app_name": app.Config.AppName,
			"pid":      os.Getpid(),
		})
		if err := fiberApp.Listen(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal(lifecycleID, "Failed to start server", map[string]any{
				"error": err.Error(),
			})
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info(lifecycleID, "Shutting down server", nil)

	// Timeout context for shutdown
	timeoutSeconds := 10
	if app.Config.Env == config.DEV {
		timeoutSeconds = 0 // No timeout in dev for easier debugging
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	if err := fiberApp.ShutdownWithContext(ctx); err != nil {
		logger.Fatal(lifecycleID, "Server forced to shutdown", map[string]any{
			"error": err.Error(),
		})
	}

	logger.Info(lifecycleID, "Server exited", nil)
}
