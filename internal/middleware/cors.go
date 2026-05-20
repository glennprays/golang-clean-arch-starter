package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORS returns CORS middleware configured for the given allowed
// origins (comma-separated). Pass cfg.CorsAllowedOrigins; the
// production guard in config.Load refuses to boot with "*".
func CORS(allowedOrigins string) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		ExposeHeaders:    "X-Request-ID",
		AllowCredentials: false,
	})
}
