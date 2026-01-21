//go:build wireinject
// +build wireinject

package infrastructure

import (
	"github.com/google/wire"

	"github.com/glennprays/golang-clean-arch-starter/config"
	"github.com/glennprays/golang-clean-arch-starter/internal/handler"
	"github.com/glennprays/golang-clean-arch-starter/internal/router"
	"github.com/glennprays/golang-clean-arch-starter/pkg/logger"
)

var CoreSet = wire.NewSet(
	config.Load,
	logger.ProviderLogger,
)

var ApiSet = wire.NewSet(
	handler.NewHealthHandler,
	router.NewRouter,
)

func InitializeApp() (*App, error) {
	wire.Build(
		CoreSet,
		ApiSet,
		wire.Struct(new(App), "*"),
	)
	return nil, nil
}
