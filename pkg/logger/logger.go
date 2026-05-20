package logger

import (
	"github.com/glennprays/golang-clean-arch-starter/config"
	"github.com/glennprays/log"
)

func ProviderLogger(config *config.Config) (*log.Logger, error) {
	return log.New(log.Config{
		Service: config.AppName,
		Env:     config.Env.String(),
		Level:   log.Level(config.LogLevel),
		Output:  log.OutputType(config.LogOutput),
	})
}
