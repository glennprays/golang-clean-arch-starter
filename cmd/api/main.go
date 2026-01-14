package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/glennprays/golang-clean-arch-starter/internal/infrastructure"
	"github.com/glennprays/golang-clean-arch-starter/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Initialize app dependencies via Wire
	app, err := infrastructure.InitializeApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	// Create Fiber app with custom error handler
	fiberApp := fiber.New(fiber.Config{
		AppName:      app.Config.AppName,
		ErrorHandler: middleware.ErrorHandler(),
	})

	// Built-in middleware
	fiberApp.Use(logger.New())
	fiberApp.Use(recover.New())

	// Setup routes (includes custom middleware)
	app.Router.Setup(fiberApp)

	// Start server in goroutine
	addr := fmt.Sprintf(":%s", app.Config.AppPort)
	go func() {
		log.Printf("Starting server on %s", addr)
		if err := fiberApp.Listen(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Timeout context for shutdown
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := fiberApp.Shutdown(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
