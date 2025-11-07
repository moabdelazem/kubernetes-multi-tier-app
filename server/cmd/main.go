package main

import (
	"net/http"

	"github.com/moabdelazem/k8s-app/internal/api"
	"github.com/moabdelazem/k8s-app/internal/config"
	"github.com/moabdelazem/k8s-app/internal/database"
	"github.com/moabdelazem/k8s-app/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Initialize configuration
	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize logger
	if err := logger.Init(cfg.Env); err != nil {
		logger.Fatal("Failed to initialize logger", zap.Error(err))
	}
	defer logger.Sync()

	// Initialize database connection
	dbConfig := &database.Config{
		Host:            cfg.DB.Host,
		Port:            cfg.DB.Port,
		User:            cfg.DB.User,
		Password:        cfg.DB.Password,
		DBName:          cfg.DB.DBName,
		SSLMode:         cfg.DB.SSLMode,
		MaxOpenConns:    cfg.DB.MaxOpenConns,
		MaxIdleConns:    cfg.DB.MaxIdleConns,
		ConnMaxLifetime: cfg.DB.ConnMaxLifetime,
		MaxRetries:      cfg.DB.MaxRetries,
		RetryDelay:      cfg.DB.RetryDelay,
	}

	if _, err := database.NewConnection(dbConfig); err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close()

	logger.Info("Database connection pool initialized",
		zap.Int("max_open_conns", cfg.DB.MaxOpenConns),
		zap.Int("max_idle_conns", cfg.DB.MaxIdleConns),
		zap.Duration("conn_max_lifetime", cfg.DB.ConnMaxLifetime),
		zap.Int("max_retries", cfg.DB.MaxRetries),
		zap.Duration("retry_delay", cfg.DB.RetryDelay),
	)

	// Setup routes with database and config
	router := api.SetupRoutes(database.GetDB(), cfg)

	// Start server
	logger.Info("Starting server",
		zap.String("address", cfg.Addr),
		zap.String("environment", cfg.Env),
	)

	if err := http.ListenAndServe(cfg.Addr, router); err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}
