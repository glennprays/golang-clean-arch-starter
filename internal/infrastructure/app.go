package infrastructure

import (
	"github.com/glennprays/golang-clean-arch-starter/config"
	"github.com/glennprays/golang-clean-arch-starter/internal/router"
	"github.com/glennprays/log"
)

type App struct {
	Config *config.Config
	Logger *log.Logger
	Router *router.Router
}
