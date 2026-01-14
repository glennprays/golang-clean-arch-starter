package main

import (
	"fmt"
	"log"

	"github.com/glennprays/golang-clean-arch-starter/internal/infrastructure"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	app, err := infrastructure.InitializeApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	fiberApp := fiber.New(fiber.Config{
		AppName: app.Config.AppName,
	})

	// Middleware
	fiberApp.Use(logger.New())
	fiberApp.Use(recover.New())

	// Routes
	fiberApp.Get("/health", app.HealthHandler.Check)

	// Start server
	addr := fmt.Sprintf(":%s", app.Config.AppPort)
	log.Printf("Starting server on %s", addr)
	if err := fiberApp.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
