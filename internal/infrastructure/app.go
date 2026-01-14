package infrastructure

import (
	"github.com/glennprays/golang-clean-arch-starter/config"
	"github.com/glennprays/golang-clean-arch-starter/internal/handler"
	"github.com/glennprays/golang-clean-arch-starter/internal/router"
)

type App struct {
	Config        *config.Config
	HealthHandler *handler.HealthHandler
	Router        *router.Router
}
