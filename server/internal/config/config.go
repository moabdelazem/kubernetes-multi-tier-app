package config

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/moabdelazem/k8s-app/pkg/env"
)

type Config struct {
	Addr string `json:"addr"`
	Env  string `json:"env"`
	DB   DBConfig
}

type DBConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	MaxRetries      int
	RetryDelay      time.Duration
}

func NewConfig() (*Config, error) {
	godotenv.Load()

	// Parse connection pool settings
	maxOpenConns, _ := strconv.Atoi(env.GetEnv("DB_MAX_OPEN_CONNS", "25"))
	maxIdleConns, _ := strconv.Atoi(env.GetEnv("DB_MAX_IDLE_CONNS", "5"))
	connMaxLifetime, _ := time.ParseDuration(env.GetEnv("DB_CONN_MAX_LIFETIME", "5m"))

	// Parse retry settings
	maxRetries, _ := strconv.Atoi(env.GetEnv("DB_MAX_RETRIES", "5"))
	retryDelay, _ := time.ParseDuration(env.GetEnv("DB_RETRY_DELAY", "2s"))

	cfg := &Config{
		Addr: fmt.Sprintf(":%s", env.GetEnv("PORT", "8080")),
		Env:  env.GetEnv("ENV", "development"),
		DB: DBConfig{
			Host:            env.GetEnv("DB_HOST", "localhost"),
			Port:            env.GetEnv("DB_PORT", "5432"),
			User:            env.GetEnv("DB_USER", "devuser"),
			Password:        env.GetEnv("DB_PASSWORD", "devpassword"),
			DBName:          env.GetEnv("DB_NAME", "k8s_app_dev"),
			SSLMode:         env.GetEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: connMaxLifetime,
			MaxRetries:      maxRetries,
			RetryDelay:      retryDelay,
		},
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
