package config

import (
	"errors"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/moabdelazem/k8s-app/pkg/env"
)

type Config struct {
	Addr string `json:"addr"`
	Env  string `json:"env"`
}

func NewConfig() (*Config, error) {
	godotenv.Load()

	cfg := &Config{
		Addr: fmt.Sprintf(":%s", env.GetEnv("PORT", "8080")),
		Env:  env.GetEnv("ENV", "development"),
	}

	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validateConfig(cfg *Config) error {
	if cfg.Addr == "" {
		return errors.New("addr is required")
	}
	if cfg.Env == "" {
		return errors.New("env is required")
	}
	return nil
}
