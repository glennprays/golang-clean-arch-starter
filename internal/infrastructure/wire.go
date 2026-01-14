//go:build wireinject
// +build wireinject

package infrastructure

import (
	"github.com/google/wire"

	"github.com/glennprays/golang-clean-arch-starter/config"
	"github.com/glennprays/golang-clean-arch-starter/internal/handler"
	"github.com/glennprays/golang-clean-arch-starter/internal/router"
)

func InitializeApp() (*App, error) {
	wire.Build(
		config.Load,
		handler.NewHealthHandler,
		router.NewRouter,
		wire.Struct(new(App), "*"),
	)
	return nil, nil
}
