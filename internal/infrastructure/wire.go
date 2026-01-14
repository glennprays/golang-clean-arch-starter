//go:build wireinject
// +build wireinject

package infrastructure

import (
	"github.com/google/wire"

	"github.com/glennprays/golang-clean-arch-starter/internal/config"
	"github.com/glennprays/golang-clean-arch-starter/internal/handler"
)

func InitializeApp() (*App, error) {
	wire.Build(
		config.Load,
		handler.NewHealthHandler,
		wire.Struct(new(App), "*"),
	)
	return nil, nil
}
