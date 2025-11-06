package main

import (
	"net/http"

	"github.com/moabdelazem/k8s-app/internal/api"
	"github.com/moabdelazem/k8s-app/internal/config"
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

	// Setup routes
	router := api.SetupRoutes()

	// Start server
	logger.Info("Starting server",
		zap.String("address", cfg.Addr),
		zap.String("environment", cfg.Env),
	)

	if err := http.ListenAndServe(cfg.Addr, router); err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}
