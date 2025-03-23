package config

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sethvargo/go-envconfig"
	"github.com/zanz1n/blog/internal/utils"
)

var config = utils.NewLazy(initConfig)

type Config struct {
	ListenAddr string `env:"LISTEN_ADDR, default=:8080"`

	DatabaseUrl string `env:"DATABASE_URL, default=file:$DATA_DIR/sqlite.db"`
	RedisUrl    string `env:"REDIS_URL"`

	LogLevel slog.Level `env:"LOG_LEVEL, default=INFO"`

	JWT JwtConfig `env:", prefix=JWT_"`
}

type JwtConfig struct {
	PrivateKey string `env:"PRIVATE_KEY, default=file:$DATA_DIR/jwt.priv.pem"`
	PublicKey  string `env:"PUBLIC_KEY, default=file:$DATA_DIR/jwt.pub.pem"`

	// In hours.
	Duration uint8 `env:"DURATION"`
}

func (c *JwtConfig) GetDuration() time.Duration {
	return time.Duration(c.Duration) * time.Hour
}

func Get() (*Config, error) {
	return config.Get()
}

func initConfig() (*Config, error) {
	var cfg Config

	err := envconfig.Process(context.Background(), &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire config: %w", err)
	}

	return &cfg, nil
}
