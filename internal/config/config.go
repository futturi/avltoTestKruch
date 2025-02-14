package config

import (
	"context"
	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"
	"testAlvtoShp/internal/logger"
)

type Config struct {
	DbPort    string `env:"DATABASE_PORT"`
	DbHost    string `env:"DATABASE_HOST"`
	DbUser    string `env:"DATABASE_USER"`
	DbPass    string `env:"DATABASE_PASSWORD"`
	DbName    string `env:"DATABASE_NAME"`
	JwtSecret string `env:"JWT_SECRET"`
	Port      string `env:"SERVER_PORT"`
}

func InitConfig(ctx context.Context) *Config {
	var cfg Config

	log := logger.LoggerFromContext(ctx)

	err := cleanenv.ReadEnv(&cfg)

	if err != nil {
		log.Fatal("error reading env config", zap.Error(err))
		return nil
	}

	return &cfg
}
