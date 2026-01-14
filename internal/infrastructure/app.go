package infrastructure

import (
	"github.com/glennprays/golang-clean-arch-starter/internal/config"
	"github.com/glennprays/golang-clean-arch-starter/internal/handler"
)

type App struct {
	Config        *config.Config
	HealthHandler *handler.HealthHandler
}
