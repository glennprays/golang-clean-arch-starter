package config

import (
	"os"

	"github.com/creasty/defaults"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Env     string `mapstructure:"ENV" default:"development"`
	AppName string `mapstructure:"APP_NAME" default:"golang-clean-architecture"`
	AppPort string `mapstructure:"APP_PORT" default:"3000"`

	DBHost     string `mapstructure:"DB_HOST" default:"localhost"`
	DBPort     string `mapstructure:"DB_PORT" default:"5432"`
	DBUser     string `mapstructure:"DB_USER" default:"postgres"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
}

func Load() (*Config, error) {
	// Check ENV explicitly - if empty, set to "development"
	if os.Getenv("ENV") == "" {
		os.Setenv("ENV", "development")
	}

	// Load .env file
	_ = godotenv.Load()

	// Configure Viper to read from environment variables
	viper.AutomaticEnv()

	// Create config instance
	cfg := &Config{}

	// Apply defaults from struct tags
	if err := defaults.Set(cfg); err != nil {
		return nil, err
	}

	// Unmarshal environment variables into config
	// This will override defaults with actual env values
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
