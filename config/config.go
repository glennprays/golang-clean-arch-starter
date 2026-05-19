package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Env     Environment `mapstructure:"ENV" default:"development" validate:"required,oneof=development staging production"`
	AppName string      `mapstructure:"APP_NAME" default:"golang-clean-architecture" validate:"required"`
	AppPort int         `mapstructure:"APP_PORT" default:"3000" validate:"required,min=1,max=65535"`

	LogLevel  string `mapstructure:"LOG_LEVEL" default:"debug" validate:"required"`
	LogOutput string `mapstructure:"LOG_OUTPUT" default:"stdout" validate:"required"`

	DBHost     string `mapstructure:"DB_HOST" default:"localhost" validate:"required"`
	DBPort     int    `mapstructure:"DB_PORT" default:"5432" validate:"required,min=1,max=65535"`
	DBUser     string `mapstructure:"DB_USER" default:"postgres" validate:"required"`
	DBPassword string `mapstructure:"DB_PASSWORD" validate:"required"`
	DBName     string `mapstructure:"DB_NAME" validate:"required"`
}

type Environment string

const (
	DEV     Environment = "development"
	STAGING Environment = "staging"
	PROD    Environment = "production"
)

func (e Environment) String() string {
	return string(e)
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Apply defaults from struct tags
	if err := defaults.Set(cfg); err != nil {
		return nil, err
	}

	envStr := strings.ToLower(os.Getenv("ENV"))
	env := Environment(envStr)
	if env == "" {
		env = DEV
	}

	// Load .env file in development
	if env == DEV {
		_ = godotenv.Load(".env")
	}

	// Configure Viper to read from environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Auto-bind each struct field by key
	t := reflect.TypeOf(cfg).Elem()
	for i := range t.NumField() {
		field := t.Field(i)
		key := field.Tag.Get("mapstructure")
		if key != "" {
			if err := viper.BindEnv(key); err != nil {
				return nil, err
			}
		}
	}

	// Unmarshal environment variables into config
	// This will override defaults with actual env values
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	// Validate required fields and enum constraints. This catches
	// missing DB_PASSWORD / DB_NAME at boot rather than at first query.
	if err := validator.New().Struct(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}
